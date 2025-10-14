package sqlite

import (
	"account-service/internal/ports"
	"gorm.io/gorm"
	"sync"
)

// TransactionRepo struct to interact with the database.
type TransactionRepo struct {
	DB *gorm.DB
	mu sync.RWMutex
}

// NewTransactionRepo creates a new TransactionRepo instance with an SQLite connection.
func NewTransactionRepo(db *gorm.DB) ports.TransactionRepo {
	return &TransactionRepo{DB: db}
}
