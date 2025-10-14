package ports

import "account-service/internal/domain/entity"

type EventRepo interface {
	CreateEvent(event *entity.Event) error
}
