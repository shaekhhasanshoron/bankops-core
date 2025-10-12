package app

import (
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
func (u *UpdateEmployee) Execute(username, newRole string) (string, error) {
	// Fetch the employee by username
	_, err := u.EmployeeRepo.GetEmployeeByUsername(username)
	if err != nil {
		logging.Logger.Warn().Err(err).Str("username", username).Msg("employee not found")
		return "", errors.New("employee not found")
	}

	// Update the role
	employee := ports.Employee{
		Username: username,
		Role:     newRole,
	}

	_, err = u.EmployeeRepo.UpdateEmployee(&employee)
	if err != nil {
		logging.Logger.Warn().Err(err).Str("username", username).Msg("failed to update employee role")
		return "", err
	}

	return "Employee role updated successfully", nil
}
