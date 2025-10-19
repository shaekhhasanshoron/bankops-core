package ports

import "transaction-service/internal/domain/entity"

type SagaRepo interface {
	CreateSaga(saga *entity.TransactionSaga) error
	GetSagaByID(id string) (*entity.TransactionSaga, error)
	GetSagaByTransactionID(transactionID string) (*entity.TransactionSaga, error)
	UpdateSaga(saga *entity.TransactionSaga) error
	GetStuckSagas() ([]*entity.TransactionSaga, error)
	GetSagasForRetry() ([]*entity.TransactionSaga, error)
	GetSagasByState(state string) ([]*entity.TransactionSaga, error)
}
