package sqlite

import (
	"account-service/internal/ports"
	"gorm.io/gorm"
	"sync"
)

// AccountRepo struct to interact with the database.
type AccountRepo struct {
	DB *gorm.DB
	mu sync.RWMutex
}

// NewAccountRepo creates a new AccountRepo instance with an SQLite connection.
func NewAccountRepo(db *gorm.DB) ports.AccountRepo {
	return &AccountRepo{DB: db}
}
