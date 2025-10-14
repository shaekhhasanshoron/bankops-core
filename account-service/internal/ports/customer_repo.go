package ports

import "account-service/internal/domain/entity"

type CustomerDTO struct {
	CustomerName string
	Requester    string
}

type CustomerRepo interface {
	CreateCustomer(customer *entity.Customer) (*entity.Customer, error)
	GetCustomerByName(name string) (*entity.Customer, error)
	GetCustomerByID(id string) (*entity.Customer, error)
	ListCustomer(page, pageSize int) ([]*entity.Customer, int64, error)
	DeleteCustomerByID(id, requester string) error
	CheckModificationAllowed(id string) error
	Exists(id string) (bool, error)
}
