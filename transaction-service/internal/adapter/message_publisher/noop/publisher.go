package noop

import (
	"context"
	"transaction-service/internal/ports"
)

// NoOpPublisher implements MessagePublisher for when messaging is disabled
type NoOpPublisher struct{}

func NewNoOpPublisher() ports.MessagePublisher {
	return &NoOpPublisher{}
}

func (n *NoOpPublisher) Publish(ctx context.Context, topic string, message string) error {
	return nil
}

func (n *NoOpPublisher) Close() error {
	return nil
}

func (n *NoOpPublisher) HealthCheck(ctx context.Context) error {
	return nil
}
