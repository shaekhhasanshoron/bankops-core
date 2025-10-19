package entity

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

const (
	EventTypeCustomerCreated      = "customer_created"
	EventTypeCustomerUpdated      = "customer_updated"
	EventTypeCustomerDeleted      = "customer_deleted"
	EventTypeAccountCreated       = "account_created"
	EventTypeAccountUpdated       = "account_updated"
	EventTypeAccountDeleted       = "account_deleted"
	EventTypeTransactionInit      = "transaction_init"
	EventTypeTransactionCommit    = "transaction_commit"
	EventTypeTransactionCompleted = "transaction_completed"
	EventTypeTransactionRollback  = "transaction_rollback"
	EventTypeTransactionFailed    = "transaction_failed"
	EventAggregateTypeCustomer    = "customer"
	EventAggregateTypeAccount     = "account"
	EventAggregateTypeTransaction = "transaction"
)

type Event struct {
	ID            string          `gorm:"primaryKey"`
	Type          string          `gorm:"not null;index"`
	AggregateID   string          `gorm:"not null;index"`
	AggregateType string          `gorm:"not null"`
	Data          json.RawMessage `gorm:"type:json"`
	Processed     bool            `gorm:"default:false"`
	Error         string          `json:"error,omitempty"`
	Version       int             `gorm:"default:1"`
	Status        string          `gorm:"not null;default:valid"`
	CreatedAt     time.Time
	CreatedBy     string `gorm:"null"`
}

func NewEvent(eventType, aggregateID, aggregateType, requester string, data interface{}) (*Event, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &Event{
		ID:            uuid.New().String(),
		Type:          eventType,
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		Data:          jsonData,
		CreatedAt:     time.Now(),
		CreatedBy:     requester,
		Version:       1,
	}, nil
}

func (e *Event) ToString() string {
	jsonData, _ := json.Marshal(&e)
	return string(jsonData)
}
