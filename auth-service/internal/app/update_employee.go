package app

import (
	"auth-service/internal/domain/entity"
	"auth-service/internal/logging"
	"auth-service/internal/ports"
	"errors"
)

// UpdateEmployee is the use-case for updating an employee's role.
type UpdateEmployee struct {
	EmployeeRepo ports.EmployeeRepo
}

// NewUpdateEmployee creates a new UpdateEmployee use-case instance.
func NewUpdateEmployee(employeeRepo ports.EmployeeRepo) *UpdateEmployee {
	return &UpdateEmployee{
		EmployeeRepo: employeeRepo,
	}
}

// Execute updates an employee's role.
func (a *UpdateEmployee) Execute(username, role, requester string) (string, error) {
	employee, err := a.EmployeeRepo.GetEmployeeByUsername(username)
	if err != nil {
		logging.Logger.Warn().Err(err).Str("username", username).Msg("employee not found")
		return "Employee not found", errors.New("employee not found")
	}

	if role == "" || (role != entity.EmployeeRoleAdmin && role != entity.EmployeeRoleViewer && role != entity.EmployeeRoleEditor) {
		logging.Logger.Warn().Err(errors.New("invalid role")).Msg("role: " + role)
		return "Invalid role", errors.New("invalid role")
	}

	employee.Role = role
	employee.UpdatedBy = requester

	_, err = a.EmployeeRepo.UpdateEmployee(employee)
	if err != nil {
		logging.Logger.Warn().Err(err).Str("username", username).Msg("failed to update employee role")
		return "Failed to update employee role", err
	}

	return "Employee role updated successfully", nil
}
