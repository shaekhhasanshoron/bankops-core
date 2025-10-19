package ports

import "transaction-service/internal/domain/entity"

type EventRepo interface {
	CreateEvent(event *entity.Event) error
}
