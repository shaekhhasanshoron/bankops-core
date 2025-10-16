package app

import (
	"auth-service/internal/domain/entity"
	"auth-service/internal/observability/metrics"
	"auth-service/internal/ports"
)

// ListEmployee is a use-case for getting account list
type ListEmployee struct {
	EmployeeRepo ports.EmployeeRepo
}

// NewListEmployee creates a new ListEmployee use-case
func NewListEmployee(employeeRepo ports.EmployeeRepo) *ListEmployee {
	return &ListEmployee{
		EmployeeRepo: employeeRepo,
	}
}

func (a *ListEmployee) Execute(page, pageSize int, sortOrder string) ([]*entity.Employee, int64, int64, string, error) {
	metrics.IncRequestActive()
	defer metrics.DecRequestActive()

	var err error
	defer func() {
		metrics.RecordOperation("list_employee", err)
	}()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 100
	}

	employee, totalCount, err := a.EmployeeRepo.ListEmployee(page, pageSize, sortOrder)

	totalPages := int64(0)
	if totalCount > 0 {
		totalPages = (totalCount + int64(pageSize) - 1) / int64(pageSize)
	}

	return employee, totalCount, totalPages, "Employee List", nil
}
