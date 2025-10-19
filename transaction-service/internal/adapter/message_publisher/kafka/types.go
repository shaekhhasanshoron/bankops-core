package kafka

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	BrokerAddr       string
	ClientID         string
	Acks             string
	Retries          int
	BatchSize        int
	LingerMs         int
	HealthCheckTopic string
}

// KafkaHealthStatus represents the health status of Kafka connection
type KafkaHealthStatus struct {
	Connected   bool     `json:"connected"`
	LastError   string   `json:"last_error,omitempty"`
	LastCheck   string   `json:"last_check"`
	BrokerCount int      `json:"broker_count"`
	Topics      []string `json:"topics,omitempty"`
}

// DefaultKafkaConfig returns default Kafka configuration
func DefaultKafkaConfig(brokerAddr string) *KafkaConfig {
	return &KafkaConfig{
		BrokerAddr:       brokerAddr,
		ClientID:         "account-service",
		Acks:             "all",
		Retries:          3,
		BatchSize:        100,
		LingerMs:         10,
		HealthCheckTopic: "__health_check",
	}
}
