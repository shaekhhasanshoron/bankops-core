package sqlite

import (
	"errors"
	"gorm.io/gorm"
	"sync"
	"time"
	"transaction-service/internal/domain/entity"
	"transaction-service/internal/ports"
)

type SagaRepo struct {
	DB *gorm.DB
	mu sync.RWMutex
}

func NewSagaRepo(db *gorm.DB) ports.SagaRepo {
	return &SagaRepo{DB: db}
}

func (r *SagaRepo) CreateSaga(saga *entity.TransactionSaga) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.DB.Create(saga).Error
}

func (r *SagaRepo) GetSagaByID(id string) (*entity.TransactionSaga, error) {
	var saga entity.TransactionSaga
	err := r.DB.Where("id = ?", id).First(&saga).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &saga, err
}

func (r *SagaRepo) GetSagaByTransactionID(transactionID string) (*entity.TransactionSaga, error) {
	var saga entity.TransactionSaga
	err := r.DB.Where("transaction_id = ?", transactionID).First(&saga).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &saga, err
}

func (r *SagaRepo) UpdateSaga(saga *entity.TransactionSaga) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	currentVersion := saga.Version
	result := r.DB.Model(saga).
		Where("id = ? AND version = ?", saga.ID, currentVersion).
		Updates(map[string]interface{}{
			"current_state":         saga.CurrentState,
			"current_step":          saga.CurrentStep,
			"compensation_required": saga.CompensationRequired,
			"compensation_reason":   saga.CompensationReason,
			"retry_count":           saga.RetryCount,
			"last_retry_at":         saga.LastRetryAt,
			"next_retry_at":         saga.NextRetryAt,
			"updated_at":            time.Now(),
			"version":               currentVersion + 1,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("concurrent modification detected")
	}

	saga.Version = currentVersion + 1
	return nil
}

func (r *SagaRepo) GetStuckSagas() ([]*entity.TransactionSaga, error) {
	var sagas []*entity.TransactionSaga
	err := r.DB.
		Where("timeout_at < ? AND current_state NOT IN (?)",
			time.Now(),
			[]string{entity.TransactionSagaStateCompleted, entity.TransactionSagaStateFailed, entity.TransactionSagaStateCompensated}).
		Find(&sagas).Error
	return sagas, err
}

func (r *SagaRepo) GetSagasForRetry() ([]*entity.TransactionSaga, error) {
	var sagas []*entity.TransactionSaga
	err := r.DB.
		Where("next_retry_at IS NOT NULL AND next_retry_at < ? AND current_state NOT IN (?)",
			time.Now(),
			[]string{entity.TransactionSagaStateCompleted, entity.TransactionSagaStateFailed, entity.TransactionSagaStateCompensated}).
		Find(&sagas).Error
	return sagas, err
}

func (r *SagaRepo) GetSagasByState(state string) ([]*entity.TransactionSaga, error) {
	var sagas []*entity.TransactionSaga
	err := r.DB.Where("current_state = ?", state).Find(&sagas).Error
	return sagas, err
}
