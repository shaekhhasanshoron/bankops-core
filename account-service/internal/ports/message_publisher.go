package ports

import "context"

// MessagePublisher defines the contract for message publishing
type MessagePublisher interface {
	Publish(ctx context.Context, topic string, message string) error
	Close() error
	HealthCheck(ctx context.Context) error
}

// MessagePublisherFactory creates message publisher instances
type MessagePublisherFactory interface {
	Create(brokerAddr string) (MessagePublisher, error)
}

// HealthCheckable interface for components that support health checks
type HealthCheckable interface {
	HealthCheck(ctx context.Context) error
}
