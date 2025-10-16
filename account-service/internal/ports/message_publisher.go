package ports

type MessagePublisher interface {
	InitConnection(brokerAddr string) error
	Publish(topic string, message string) error
	CloseConnection() error
}
