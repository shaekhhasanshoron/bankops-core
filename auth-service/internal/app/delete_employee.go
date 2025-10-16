package app

import (
	"auth-service/internal/logging"
	"auth-service/internal/ports"
	"errors"
)

// DeleteEmployee is the use-case for deleting an employee (marking as invalid).
type DeleteEmployee struct {
	EmployeeRepo ports.EmployeeRepo
}

// NewDeleteEmployee creates a new DeleteEmployee use-case instance.
func NewDeleteEmployee(employeeRepo ports.EmployeeRepo) *DeleteEmployee {
	return &DeleteEmployee{
		EmployeeRepo: employeeRepo,
	}
}

// Execute marks an employee as invalid (soft delete).
func (a *DeleteEmployee) Execute(username, requester string) (string, error) {
	// Fetch the employee by username
	_, err := a.EmployeeRepo.GetEmployeeByUsername(username)
	if err != nil {
		logging.Logger.Warn().Err(err).Str("username", username).Msg("employee not found")
		return "Employee not found", errors.New("employee not found")
	}

	err = a.EmployeeRepo.DeleteEmployee(username, requester)
	if err != nil {
		logging.Logger.Warn().Err(err).Str("username", username).Msg("failed to mark employee as invalid")
		return "Failed to delete employee", err
	}

	return "Employee deleted successfully", nil
}
