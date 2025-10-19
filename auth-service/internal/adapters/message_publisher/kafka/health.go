package kafka

import (
	"auth-service/internal/logging"
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"time"
)

// HealthMonitor manages Kafka connection health monitoring and automatic reconnection
type HealthMonitor struct {
	publisher            *KafkaPublisher
	checkInterval        time.Duration
	retryInterval        time.Duration
	stopChan             chan struct{}
	running              bool
	reconnectAttempts    int
	maxReconnectAttempts int
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(publisher *KafkaPublisher, checkInterval, retryInterval time.Duration) *HealthMonitor {
	return &HealthMonitor{
		publisher:            publisher,
		checkInterval:        checkInterval,
		retryInterval:        retryInterval,
		stopChan:             make(chan struct{}),
		maxReconnectAttempts: 0,
	}
}

// Start begins the health monitoring process
func (hm *HealthMonitor) Start() {
	if hm.running {
		return
	}

	hm.running = true
	go hm.monitorLoop()

	logging.Logger.Info().
		Str("check_interval", hm.checkInterval.String()).
		Str("retry_interval", hm.retryInterval.String()).
		Msg("Kafka health monitor started")
}

// Stop stops the health monitoring process
func (hm *HealthMonitor) Stop() {
	if !hm.running {
		return
	}

	close(hm.stopChan)
	hm.running = false

	logging.Logger.Info().Msg("Kafka health monitor stopped")
}

// monitorLoop continuously monitors Kafka health
func (hm *HealthMonitor) monitorLoop() {
	// Do an immediate health check on start
	hm.performHealthCheck()

	ticker := time.NewTicker(hm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-hm.stopChan:
			return
		case <-ticker.C:
			hm.performHealthCheck()
		}
	}
}

// performHealthCheck performs a single health check and handles reconnection
func (hm *HealthMonitor) performHealthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := hm.publisher.HealthCheck(ctx); err != nil {
		logging.Logger.Warn().
			Err(err).
			Str("broker_addr", hm.publisher.config.BrokerAddr).
			Int("reconnect_attempts", hm.reconnectAttempts).
			Msg("Kafka health check failed")

		// If we're not connected, try to reestablish connection
		if !hm.publisher.IsConnected() {
			hm.attemptReconnection()
		}
	} else {
		if hm.reconnectAttempts > 0 {
			logging.Logger.Info().
				Str("broker_addr", hm.publisher.config.BrokerAddr).
				Msg("Kafka connection restored")
			hm.reconnectAttempts = 0 // Reset counter on successful connection
		} else {
			logging.Logger.Debug().
				Str("broker_addr", hm.publisher.config.BrokerAddr).
				Msg("Kafka health check passed")
		}
	}
}

// attemptReconnection tries to reestablish Kafka connection
func (hm *HealthMonitor) attemptReconnection() {
	// Check if we've exceeded maximum reconnection attempts
	if hm.maxReconnectAttempts > 0 && hm.reconnectAttempts >= hm.maxReconnectAttempts {
		logging.Logger.Error().
			Str("broker_addr", hm.publisher.config.BrokerAddr).
			Int("attempts", hm.reconnectAttempts).
			Int("max_attempts", hm.maxReconnectAttempts).
			Msg("Maximum Kafka reconnection attempts exceeded")
		return
	}

	hm.reconnectAttempts++

	logging.Logger.Info().
		Str("broker_addr", hm.publisher.config.BrokerAddr).
		Int("attempt", hm.reconnectAttempts).
		Msg("Attempting to reconnect to Kafka")

	// Close existing connection if any
	hm.publisher.Close()

	// Retry logic
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case <-hm.stopChan:
			return
		default:
			// Create new producer
			producer, err := kafka.NewProducer(&kafka.ConfigMap{
				"bootstrap.servers": hm.publisher.config.BrokerAddr,
				"client.id":         hm.publisher.config.ClientID,
			})

			if err != nil {
				logging.Logger.Warn().
					Err(err).
					Int("attempt", attempt).
					Int("max_retries", maxRetries).
					Msg("Failed to recreate Kafka producer")

				if attempt < maxRetries {
					time.Sleep(hm.retryInterval * time.Duration(attempt))
				}
				continue
			}

			// Update publisher with new producer
			hm.publisher.mu.Lock()
			hm.publisher.producer = producer
			hm.publisher.mu.Unlock()

			// Test the new connection
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err = hm.publisher.HealthCheck(ctx)
			cancel()

			if err == nil {
				logging.Logger.Info().
					Str("broker_addr", hm.publisher.config.BrokerAddr).
					Int("reconnect_attempts", hm.reconnectAttempts).
					Msg("Successfully reconnected to Kafka")
				return
			}

			hm.publisher.Close()
		}
	}

	logging.Logger.Warn().
		Str("broker_addr", hm.publisher.config.BrokerAddr).
		Int("reconnect_attempts", hm.reconnectAttempts).
		Msg("Failed to reconnect to Kafka, will retry on next health check")
}
