package app

import (
	"auth-service/internal/common"
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
func (u *UpdateEmployee) Execute(username, role string) (string, error) {
	_, err := u.EmployeeRepo.GetEmployeeByUsername(username)
	if err != nil {
		logging.Logger.Warn().Err(err).Str("username", username).Msg("employee not found")
		return "Employee not found", errors.New("employee not found")
	}

	if role == "" || (role != common.EmployeeRoleAdmin && role != common.EmployeeRoleViewer && role != common.EmployeeRoleEditor) {
		logging.Logger.Warn().Err(errors.New("invalid role")).Msg("role: " + role)
		return "Invalid role", errors.New("invalid role")
	}

	// Update the role
	employee := ports.Employee{
		Username: username,
		Role:     role,
	}

	_, err = u.EmployeeRepo.UpdateEmployee(&employee)
	if err != nil {
		logging.Logger.Warn().Err(err).Str("username", username).Msg("failed to update employee role")
		return "Failed to update employee role", err
	}

	return "Employee role updated successfully", nil
}
