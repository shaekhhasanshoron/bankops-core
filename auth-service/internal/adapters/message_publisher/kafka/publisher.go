package kafka

import (
	"auth-service/internal/logging"
	"auth-service/internal/ports"
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"sync"
	"time"
)

// KafkaPublisher implements the MessagePublisher interface with health monitoring
type KafkaPublisher struct {
	producer        *kafka.Producer
	config          *KafkaConfig
	mu              sync.RWMutex
	isConnected     bool
	lastError       error
	lastHealthCheck time.Time
	metadata        *kafka.Metadata
}

// NewKafkaPublisher creates a new Kafka publisher factory
func NewKafkaPublisher() ports.MessagePublisherFactory {
	return &kafkaFactory{}
}

type kafkaFactory struct{}

func (kf *kafkaFactory) Create(brokerAddr string) (ports.MessagePublisher, error) {
	config := DefaultKafkaConfig(brokerAddr)

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":  config.BrokerAddr,
		"client.id":          config.ClientID,
		"acks":               config.Acks,
		"retries":            config.Retries,
		"batch.size":         config.BatchSize,
		"linger.ms":          config.LingerMs,
		"enable.idempotence": true,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	publisher := &KafkaPublisher{
		producer: producer,
		config:   config,
	}

	// Perform initial health check
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := publisher.HealthCheck(ctx); err != nil {
		logging.Logger.Warn().Err(err).Msg("Initial Kafka health check failed")
	}

	return publisher, nil
}

// Publish sends a message to Kafka with context support
func (kp *KafkaPublisher) Publish(ctx context.Context, topic string, message string) error {
	kp.mu.RLock()
	defer kp.mu.RUnlock()

	if !kp.isConnected {
		return fmt.Errorf("kafka publisher not connected")
	}

	deliveryChan := make(chan kafka.Event)
	defer close(deliveryChan)

	err := kp.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: []byte(message),
	}, deliveryChan)

	if err != nil {
		kp.setConnectionStatus(false, err)
		return fmt.Errorf("failed to produce message: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-deliveryChan:
		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			kp.setConnectionStatus(false, m.TopicPartition.Error)
			return fmt.Errorf("delivery failed: %w", m.TopicPartition.Error)
		}
		kp.setConnectionStatus(true, nil)
		return nil
	}
}

// HealthCheck performs a health check on the Kafka connection
func (kp *KafkaPublisher) HealthCheck(ctx context.Context) error {
	kp.mu.Lock()
	defer kp.mu.Unlock()

	if kp.producer == nil {
		return fmt.Errorf("kafka producer is nil")
	}

	// Get cluster metadata with timeout
	metadata, err := kp.producer.GetMetadata(nil, true, 5000) // 5 second timeout
	if err != nil {
		kp.setConnectionStatus(false, err)
		return fmt.Errorf("failed to get Kafka metadata: %w", err)
	}

	kp.metadata = metadata
	kp.setConnectionStatus(true, nil)
	kp.lastHealthCheck = time.Now()

	logging.Logger.Debug().
		Int("broker_count", len(metadata.Brokers)).
		Int("topic_count", len(metadata.Topics)).
		Msg("Kafka health check successful")

	return nil
}

// GetHealthStatus returns detailed health status
func (kp *KafkaPublisher) GetHealthStatus() *KafkaHealthStatus {
	kp.mu.RLock()
	defer kp.mu.RUnlock()

	status := &KafkaHealthStatus{
		Connected: kp.isConnected,
		LastCheck: kp.lastHealthCheck.Format(time.RFC3339),
	}

	if kp.lastError != nil {
		status.LastError = kp.lastError.Error()
	}

	if kp.metadata != nil {
		status.BrokerCount = len(kp.metadata.Brokers)

		topics := make([]string, 0, len(kp.metadata.Topics))
		for topic := range kp.metadata.Topics {
			topics = append(topics, topic)
		}
		status.Topics = topics
	}

	return status
}

// Close closes the Kafka producer
func (kp *KafkaPublisher) Close() error {
	kp.mu.Lock()
	defer kp.mu.Unlock()

	if kp.producer != nil {
		kp.producer.Close()
		kp.producer = nil
	}

	kp.isConnected = false
	return nil
}

// setConnectionStatus updates the connection status
func (kp *KafkaPublisher) setConnectionStatus(connected bool, err error) {
	kp.isConnected = connected
	kp.lastError = err
	if connected {
		kp.lastHealthCheck = time.Now()
	}
}

// IsConnected returns the connection status
func (kp *KafkaPublisher) IsConnected() bool {
	kp.mu.RLock()
	defer kp.mu.RUnlock()
	return kp.isConnected
}
