package message_publisher

import (
	kafka_adapter "account-service/internal/adapters/message_publisher/kafka"
	"account-service/internal/config"
	"account-service/internal/logging"
	"account-service/internal/ports"
	"encoding/json"
	"fmt"
	"sync"
)

var globalOnce sync.Once
var globalPublisherPort ports.MessagePublisher
var globalPublisherConnInitiated bool

type MessagePublishRequest struct {
	Message string
	Status  bool
}

func Init() error {
	if !config.Current().MessagePublisher.Enabled {
		logging.Logger.Info().Msg("message publish disabled; using no-op message publish to broker")
		return nil
	}

	if config.Current().MessagePublisher.BrokerType == config.BrokerTypeKafka {
		messagePublisherAdapter := kafka_adapter.NewMessagePublisher()
		err := messagePublisherAdapter.InitConnection(config.Current().MessagePublisher.BrokerAddr)
		if err != nil {
			logging.Logger.Warn().Err(err).Str("broker_addr", config.Current().MessagePublisher.BrokerAddr).
				Msg("Unable to initialize Message Broker (KAFKA)")
			return err
		}
		logging.Logger.Info().Str("broker_addr", config.Current().MessagePublisher.BrokerAddr).
			Msg("Initiated Message Broker (KAFKA)")
		setGlobal(messagePublisherAdapter, true)
	}
	return nil
}

func setGlobal(kp ports.MessagePublisher, connected bool) {
	globalOnce.Do(func() {
		if connected {
			globalPublisherPort = kp
		}
		globalPublisherConnInitiated = connected
	})
}

func getPublisher() ports.MessagePublisher {
	return globalPublisherPort
}

func Publish(request MessagePublishRequest) error {
	if !config.Current().MessagePublisher.Enabled {
		logging.Logger.Info().Msg("message publish disabled; using no-op message publish to broker")
		return nil
	}

	if globalPublisherConnInitiated {
		message, _ := json.Marshal(request)
		err := getPublisher().Publish(config.Current().MessagePublisher.PublishTopic, string(message))
		if err != nil {
			logging.Logger.Warn().Err(err).Msg("Unable to publish message")
			return fmt.Errorf("connection not stablished")
		} else {
			logging.Logger.Info().Str("topic", config.Current().MessagePublisher.PublishTopic).
				Msg("Message published successfully to publisher")
			return nil
		}
	} else {
		logging.Logger.Warn().Err(fmt.Errorf("connection not stablished")).Msg("Message not published")
		return fmt.Errorf("connection not stablished")
	}
}

func CloseConnection() error {
	if globalPublisherConnInitiated {
		return getPublisher().CloseConnection()
	}
	return nil
}

//func initKafkaBroker() error {
//	p, err := kafka.NewProducer(&kafka.ConfigMap{
//		"bootstrap.servers": config.Current().MessagePublisher.BrokerAddr,
//	})
//
//	if err != nil {
//		logging.Logger.Warn().Err(err).Str("broker", "kafka").
//			Str("broker_addr", config.Current().MessagePublisher.BrokerAddr).
//			Msg("Failed to inti kafka producer")
//
//		setGlobal(nil, true)
//	} else {
//		logging.Logger.Info().Str("broker", "kafka").
//			Str("broker_addr", config.Current().MessagePublisher.BrokerAddr).
//			Msg("Kafka Producer Initialized")
//
//		mp := kafka_adapter.NewMessagePublisher(p)
//		setGlobal(&mp, true)
//	}
//	return nil
//}
