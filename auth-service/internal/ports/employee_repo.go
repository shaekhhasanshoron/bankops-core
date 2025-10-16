package ports

import "auth-service/internal/domain/entity"

// EmployeeRepo defines the interface for employee-related database operations
type EmployeeRepo interface {
	CreateEmployee(input *entity.Employee) (*entity.Employee, error)
	GetEmployeeByUsername(username string) (*entity.Employee, error)
	UpdateEmployee(employee *entity.Employee) (*entity.Employee, error)
	DeleteEmployee(username, requester string) error
	ListEmployee(page, pageSize int, sortOrder string) ([]*entity.Employee, int64, error)
}
