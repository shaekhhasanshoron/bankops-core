package sqlite

import (
	"account-service/internal/domain/entity"
	"account-service/internal/domain/value"
	"account-service/internal/ports"
	"errors"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"gorm.io/gorm"
	"sync"
)

// CustomerRepo struct to interact with the database.
type CustomerRepo struct {
	DB *gorm.DB
	mu sync.RWMutex
}

// NewCustomerRepo creates a new CustomerRepo instance with an SQLite connection.
func NewCustomerRepo(db *gorm.DB) ports.CustomerRepo {
	return &CustomerRepo{DB: db}
}

// CreateCustomer checks if the customer already exists with a valid status and creates a new one
func (r *CustomerRepo) CreateCustomer(input *ports.CustomerDTO) (*entity.Customer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var employee entity.Customer

	// Checks if customer exists with valid status
	if err := r.DB.Where("name = ? AND status = ?", input.CustomerName, entity.CustomerStatusValid).First(&employee).Error; err == nil {
		return nil, errors.New("customer with this username already exists")
	}

	newCustomer, err := entity.NewCustomer(input.CustomerName, input.Requester)
	if err != nil {
		return nil, err
	}

	if err := r.DB.Create(newCustomer).Error; err != nil {
		return nil, err
	}

	return newCustomer, nil
}

func (r *CustomerRepo) GetCustomerByID(id string) (*entity.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var customer entity.Customer
	err := r.DB.Preload("Accounts").Where("id = ? AND status = ?", id, entity.CustomerStatusValid).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, err
}

func (r *CustomerRepo) GetCustomerByName(name string) (*entity.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var customer entity.Customer
	err := r.DB.Preload("Accounts").Where("name = ? AND status = ?", name, entity.CustomerStatusValid).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, err
}

func (r *CustomerRepo) ListCustomer(page, pageSize int) ([]*entity.Customer, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var customers []*entity.Customer
	var total int64

	offset := (page - 1) * pageSize

	if err := r.DB.Model(&entity.Customer{}).
		Where("status = ?", entity.CustomerStatusValid).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Query only valid customers
	err := r.DB.Where("status = ?", entity.CustomerStatusValid).
		Preload("Accounts").
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&customers).Error

	return customers, total, err
}

func (r *CustomerRepo) DeleteCustomerByID(id string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.DB.Where("id = ? AND status = ?", id, entity.CustomerStatusValid).Delete(&entity.Customer{}).Error
}

func (r *CustomerRepo) CheckModificationAllowed(id string) error {
	customer, err := r.GetCustomerByID(id)
	if err != nil {
		return err
	}
	if customer == nil {
		return value.ErrCustomerNotFound
	}

	if customer.LockedForOperation {
		return value.ErrCustomerLocked
	}

	if customer.HasAccountsInTransaction() {
		return value.ErrCustomerHasActiveTx
	}

	return nil
}
