package sqlite

import (
	"account-service/internal/domain/entity"
	"account-service/internal/ports"
	"gorm.io/gorm"
	"sync"
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
