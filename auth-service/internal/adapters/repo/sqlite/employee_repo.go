package sqlite

import (
	"auth-service/internal/domain/entity"
	"auth-service/internal/ports"
	"errors"
	"gorm.io/gorm"
	"sync"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// EmployeeRepo struct to interact with the database.
type EmployeeRepo struct {
	DB *gorm.DB
	mu sync.RWMutex
}

// NewEmployeeRepo creates a new EmployeeRepo instance with an SQLite connection.
func NewEmployeeRepo(db *gorm.DB) ports.EmployeeRepo {
	return &EmployeeRepo{DB: db}
}

// CreateEmployee checks if the user already exists with a valid status and creates a new one
func (r *EmployeeRepo) CreateEmployee(employee *entity.Employee) (*entity.Employee, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var existingEmployee entity.Employee

	if err := r.DB.Where("username = ? AND status = ?", employee.Username, entity.EmployeeStatusValid).First(&existingEmployee).Error; err == nil {
		return nil, errors.New("employee with this username already exists")
	}

	if err := r.DB.Create(employee).Error; err != nil {
		return nil, err
	}

	return employee, nil
}

// GetEmployeeByUsername returns the employee if valid
func (r *EmployeeRepo) GetEmployeeByUsername(username string) (*entity.Employee, error) {
	var employee entity.Employee
	if err := r.DB.Where("username = ? AND status = ?", username, entity.EmployeeStatusValid).First(&employee).Error; err != nil {
		return nil, err
	}
	return &employee, nil
}

// UpdateEmployee updates the role for a valid employee
func (r *EmployeeRepo) UpdateEmployee(input *entity.Employee) (*entity.Employee, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var employee entity.Employee
	if err := r.DB.Where("username = ? AND status = ?", input.Username, entity.EmployeeStatusValid).First(&employee).Error; err != nil {
		return nil, err
	}

	if err := r.DB.Save(&employee).Error; err != nil {
		return nil, err
	}

	return &employee, nil
}

// DeleteEmployee marks an employee as invalid (soft delete)
func (r *EmployeeRepo) DeleteEmployee(username, requester string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var employee entity.Employee
	if err := r.DB.Where("username = ? AND status = ?", username, entity.EmployeeStatusValid).First(&employee).Error; err != nil {
		return err
	}

	// Mark as invalid instead of deleting
	employee.ActiveStatus = entity.EmployeeActiveStatusDeactivated
	employee.Status = entity.EmployeeStatusInvalid
	employee.UpdatedBy = requester
	if err := r.DB.Save(&employee).Error; err != nil {
		return err
	}

	return nil
}

func (r *EmployeeRepo) ListEmployee(page, pageSize int, sortOrder string) ([]*entity.Employee, int64, error) {
	var employees []*entity.Employee
	var total int64

	offset := (page - 1) * pageSize

	if err := r.DB.Model(&entity.Employee{}).
		Where("status = ?", entity.EmployeeStatusValid).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "created_at DESC"
	if sortOrder == "asc" {
		order = "created_at ASC"
	}

	err := r.DB.Where("status = ?", entity.EmployeeStatusValid).
		Offset(offset).
		Limit(pageSize).
		Order(order).
		Find(&employees).Error

	return employees, total, err
}
