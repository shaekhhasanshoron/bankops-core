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
func (r *EmployeeRepo) CreateEmployee(input *ports.Employee) (*entity.Employee, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var employee entity.Employee

	// Checks if employee exists with valid status
	if err := r.DB.Where("username = ? AND status = ?", input.Username, entity.EmployeeStatusValid).First(&employee).Error; err == nil {
		return nil, errors.New("employee with this username already exists")
	}

	newEmployee, err := entity.NewEmployee(
		input.Username,
		input.Password,
		input.Role,
		"",
		input.Requester,
	)

	if err != nil {
		return nil, err
	}

	if err := r.DB.Create(newEmployee).Error; err != nil {
		return nil, err
	}

	return newEmployee, nil
}

// GetEmployeeByUsername returns the employee if valid
func (r *EmployeeRepo) GetEmployeeByUsername(username string) (*entity.Employee, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var employee entity.Employee
	if err := r.DB.Where("username = ? AND status = ?", username, entity.EmployeeStatusValid).First(&employee).Error; err != nil {
		return nil, err
	}
	return &employee, nil
}

// UpdateEmployee updates the role for a valid employee
func (r *EmployeeRepo) UpdateEmployee(input *ports.Employee) (*entity.Employee, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var employee entity.Employee
	if err := r.DB.Where("username = ? AND status = ?", input.Username, entity.EmployeeStatusValid).First(&employee).Error; err != nil {
		return nil, err
	}

	// Update role
	employee.Role = input.Role
	employee.UpdatedBy = input.Requester
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
