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
func (r *CustomerRepo) CreateCustomer(customer *entity.Customer) (*entity.Customer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Checks if customer exists with valid status
	if err := r.DB.Where("name = ? AND status = ?", customer.Name, entity.CustomerStatusValid).First(&entity.Customer{}).Error; err == nil {
		return nil, errors.New("customer with this username already exists")
	}

	if err := r.DB.Create(customer).Error; err != nil {
		return nil, err
	}

	return customer, nil
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
	var customer entity.Customer
	err := r.DB.Preload("Accounts").Where("name = ? AND status = ?", name, entity.CustomerStatusValid).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, err
}

func (r *CustomerRepo) ListCustomer(page, pageSize int) ([]*entity.Customer, int64, error) {
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
		//Preload("Accounts", "status = ?", entity.AccountStatusValid).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&customers).Error

	return customers, total, err
}

func (r *CustomerRepo) DeleteCustomerByID(id, requester string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	var customer entity.Customer
	if err := r.DB.Where("id = ? AND status = ?", id, entity.CustomerStatusValid).First(&customer).Error; err != nil {
		return err
	}

	customer.Status = entity.CustomerStatusInvalid
	customer.UpdatedBy = requester
	if err := r.DB.Save(&customer).Error; err != nil {
		return err
	}
	return nil
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

func (r *CustomerRepo) Exists(id string) (bool, error) {
	var count int64
	err := r.DB.Model(&entity.Customer{}).Where("id = ? AND status = ?", id, entity.CustomerStatusValid).Count(&count).Error
	return count > 0, err
}
