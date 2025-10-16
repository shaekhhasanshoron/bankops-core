package kafka

import (
	"account-service/internal/ports"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"sync"
)

// KafkaMessagePublisher struct to interact with the message publisher.
type KafkaMessagePublisher struct {
	Producer *kafka.Producer
	mu       sync.RWMutex
}

// NewMessagePublisher creates a new KafkaMessagePublisher instance with a Kafka connection.
func NewMessagePublisher() ports.MessagePublisher {
	return &KafkaMessagePublisher{}
}

func (km *KafkaMessagePublisher) InitConnection(brokerAddr string) error {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokerAddr,
	})

	if err != nil {
		return err
	} else {
		km.Producer = p
		return nil
	}
}

func (km *KafkaMessagePublisher) Publish(topic string, message string) error {
	deliveryChan := make(chan kafka.Event)
	err := km.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(message),
	}, deliveryChan)

	if err != nil {
		return err
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		return fmt.Errorf("delivery failed: %v", m.TopicPartition.Error)
	}
	close(deliveryChan)
	return m.TopicPartition.Error
}

func (km *KafkaMessagePublisher) CloseConnection() error {
	if km.Producer != nil {
		km.Producer.Close()
	}
	return nil
}
