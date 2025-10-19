package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"transaction-service/internal/adapter/message_publisher/kafka"
	"transaction-service/internal/adapter/message_publisher/noop"
	"transaction-service/internal/config"
	"transaction-service/internal/logging"
	"transaction-service/internal/ports"
)

const (
	MessageConnectionTypeKafka        = "kafka"
	MessageConnectionTypeNoOp         = "noop"
	MessageConnectionTypeDisconnected = "disconnected"

	MessageTypeTransactionCompleted = "TransactionCompleted"
	MessageTypeTransactionFailed    = "TransactionFailed"
)

type Service struct {
	publisher      ports.MessagePublisher
	healthMonitor  *kafka.HealthMonitor
	config         *config.MessagePublisherConfig
	mu             sync.RWMutex
	initialized    bool
	enabled        bool
	connectionType string
}

type Message struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Status  bool   `json:"status"`
	SentAt  string `json:"sent_at"`
}

var (
	instance *Service
	once     sync.Once
)

// GetService returns the singleton messaging service instance
func GetService() *Service {
	once.Do(func() {
		instance = &Service{}
	})
	return instance
}

// SetupMessaging initializes the messaging service without failing if connection fails
func SetupMessaging() {
	err := GetService().Initialize()
	if err != nil {
		logging.Logger.Error().
			Err(err).
			Msg("Failed to initialize messaging service due to configuration error")
		return
	}

	GetService().LogMessagingStatus()

	logging.Logger.Info().
		Bool("enabled", GetService().IsEnabled()).
		Str("connection_type", GetService().GetConnectionType()).
		Msg("Messaging service initialization completed")
}

// LogMessagingStatus logs the current messaging service status for debugging
func (s *Service) LogMessagingStatus() {
	status := GetService().GetHealthStatus()

	if GetService().IsEnabled() {
		if GetService().IsConnected() {
			logging.Logger.Info().
				Interface("messaging_status", status).
				Msg("Messaging service is enabled and connected")
		} else {
			logging.Logger.Warn().
				Interface("messaging_status", status).
				Msg("Messaging service is enabled but currently disconnected - health monitor will attempt reconnection")
		}
	} else {
		logging.Logger.Info().
			Interface("messaging_status", status).
			Msg("Messaging service is disabled")
	}
}

// Initialize sets up the messaging service but doesn't fail if connection fails
func (s *Service) Initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cfg := config.Current().MessagePublisher

	s.config = &cfg
	s.enabled = cfg.Enabled

	if !cfg.Enabled {
		s.publisher = noop.NewNoOpPublisher()
		s.initialized = true
		s.connectionType = "noop"
		logging.Logger.Info().Msg("messaging disabled; using no-op publisher")
		return nil
	}

	switch cfg.BrokerType {
	case config.BrokerTypeKafka:
		s.initializeKafka(&cfg)
	default:
		logging.Logger.Warn().
			Str("broker_type", cfg.BrokerType).
			Msg("unsupported broker type; using no-op publisher")
		s.publisher = noop.NewNoOpPublisher()
		s.connectionType = "noop"
	}

	s.initialized = true
	logging.Logger.Info().
		Str("connection_type", s.connectionType).
		Bool("enabled", s.enabled).
		Msg("messaging service initialized")

	return nil
}

// initializeKafka attempts to initialize Kafka connection but falls back to disconnected mode if it fails
func (s *Service) initializeKafka(cfg *config.MessagePublisherConfig) {
	factory := kafka.NewKafkaPublisher()
	publisher, err := factory.Create(cfg.BrokerAddr)
	if err != nil {
		logging.Logger.Warn().Err(err).
			Str("broker_addr", cfg.BrokerAddr).
			Msg("failed to initialize Kafka publisher; will retry via health monitor")

		disconnectedPublisher := &kafka.KafkaPublisher{}
		s.publisher = disconnectedPublisher
		s.connectionType = "disconnected"

		s.healthMonitor = kafka.NewHealthMonitor(
			disconnectedPublisher,
			30*time.Second,
			10*time.Second,
		)
		s.healthMonitor.Start()

		logging.Logger.Info().
			Str("broker_addr", cfg.BrokerAddr).
			Msg("Kafka health monitor started for reconnection attempts")
		return
	}

	s.publisher = publisher
	s.connectionType = "kafka"

	// Start health monitoring for Kafka
	if kafkaPublisher, ok := publisher.(*kafka.KafkaPublisher); ok {
		s.healthMonitor = kafka.NewHealthMonitor(
			kafkaPublisher,
			30*time.Second,
			10*time.Second,
		)
		s.healthMonitor.Start()

		logging.Logger.Info().
			Str("broker_addr", cfg.BrokerAddr).
			Msg("Kafka publisher initialized with health monitoring")
	}
}

// Publish sends a message to the specified topic with context support
func (s *Service) Publish(topic string, message Message) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if !s.initialized {
		return fmt.Errorf("messaging service not initialized")
	}

	if !s.enabled {
		logging.Logger.Debug().Msg("messaging disabled; skipping publish")
		return nil
	}

	if s.connectionType == "disconnected" {
		logging.Logger.Warn().
			Str("topic", topic).
			Msg("messaging service is disconnected; message not published")
		return fmt.Errorf("messaging service is currently disconnected")
	}

	if message.SentAt == "" {
		message.SentAt = time.Now().Format(time.RFC3339)
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = s.publisher.Publish(ctx, topic, string(messageBytes))
	if err != nil {
		logging.Logger.Warn().Err(err).
			Str("topic", topic).
			Msg("failed to publish message")
		return fmt.Errorf("failed to publish message: %w", err)
	}

	logging.Logger.Debug().
		Str("topic", topic).
		Msg("message published successfully")
	return nil
}

// PublishToDefaultTopic publishes a message to the configured default topic
func (s *Service) PublishToDefaultTopic(message Message) error {
	if s.config == nil {
		return fmt.Errorf("messaging service not configured")
	}
	return s.Publish(s.config.PublishTopic, message)
}

// HealthCheck performs a health check on the messaging service
func (s *Service) HealthCheck(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.initialized {
		return fmt.Errorf("messaging service not initialized")
	}

	if !s.enabled {
		return nil
	}

	if s.publisher == nil {
		return fmt.Errorf("message publisher is nil")
	}

	if s.connectionType == "disconnected" {
		return fmt.Errorf("messaging service is disconnected from broker")
	}

	return s.publisher.HealthCheck(ctx)
}

// GetHealthStatus returns detailed health status
func (s *Service) GetHealthStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := map[string]interface{}{
		"enabled":         s.enabled,
		"initialized":     s.initialized,
		"connection_type": s.connectionType,
	}

	if s.enabled && s.config != nil {
		status["broker_type"] = s.config.BrokerType
		status["broker_addr"] = s.config.BrokerAddr
	}

	if s.connectionType == "kafka" {
		if kafkaPublisher, ok := s.publisher.(*kafka.KafkaPublisher); ok {
			kafkaStatus := kafkaPublisher.GetHealthStatus()
			status["kafka"] = kafkaStatus
			status["healthy"] = kafkaStatus.Connected
		}
	} else if s.connectionType == "disconnected" {
		status["healthy"] = false
		status["message"] = "disconnected from broker - health monitor attempting reconnection"
	} else {
		status["healthy"] = true
	}

	return status
}

// Close closes the messaging service and stops health monitoring
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.healthMonitor != nil {
		s.healthMonitor.Stop()
	}

	if s.publisher != nil {
		return s.publisher.Close()
	}

	s.initialized = false
	return nil
}

// IsEnabled returns whether messaging is enabled
func (s *Service) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// IsInitialized returns whether the service is initialized
func (s *Service) IsInitialized() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.initialized
}

// IsConnected returns whether the service is connected to the message broker
func (s *Service) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connectionType == MessageConnectionTypeKafka || s.connectionType == MessageConnectionTypeNoOp
}

// GetConnectionType returns the current connection type
func (s *Service) GetConnectionType() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connectionType
}
