package ports

import "auth-service/internal/domain/entity"

// Employee represents an employee in the system
type Employee struct {
	Username  string
	Password  string
	Role      string
	Requester string
}

// EmployeeRepo defines the interface for employee-related database operations
type EmployeeRepo interface {
	CreateEmployee(employee *Employee) (*entity.Employee, error)
	GetEmployeeByUsername(username string) (*entity.Employee, error)
	UpdateEmployee(employee *Employee) (*entity.Employee, error)
	DeleteEmployee(username, requester string) error
}
