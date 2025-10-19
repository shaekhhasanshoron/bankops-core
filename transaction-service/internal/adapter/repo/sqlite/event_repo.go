package sqlite

import (
	"gorm.io/gorm"
	"sync"
	"transaction-service/internal/domain/entity"
	"transaction-service/internal/ports"
)

// EventRepo struct to interact with the database.
type EventRepo struct {
	DB *gorm.DB
	mu sync.RWMutex
}

// NewEventRepo creates a new EventRepo instance with an SQLite connection.
func NewEventRepo(db *gorm.DB) ports.EventRepo {
	return &EventRepo{DB: db}
}

func (r *EventRepo) CreateEvent(event *entity.Event) error {
	return r.DB.Create(event).Error
}
