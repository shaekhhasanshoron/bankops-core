package app

import (
	"auth-service/internal/domain/entity"
	custom_err "auth-service/internal/domain/error"
	"auth-service/internal/logging"
	"auth-service/internal/observability/metrics"
	"auth-service/internal/ports"
	"errors"
	"strings"
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
	metrics.IncRequestActive()
	defer metrics.DecRequestActive()

	var err error
	defer func() {
		metrics.RecordOperation("update_employee", err)
	}()

	username = strings.TrimSpace(username)
	role = strings.TrimSpace(role)

	if username == "" || role == "" {
		logging.Logger.Warn().Msg("Invalid request")
		return "Missing required data username and role", custom_err.ErrMissingRequiredData
	}

	if role != entity.EmployeeRoleAdmin && role != entity.EmployeeRoleViewer && role != entity.EmployeeRoleEditor {
		logging.Logger.Warn().Err(custom_err.ErrInvalidRole).Str("role", role).Msg("Invalid request")
		return "Invalid role", custom_err.ErrInvalidRole
	}

	employee, err := a.EmployeeRepo.GetEmployeeByUsername(username)
	if err != nil {
		logging.Logger.Warn().Err(err).Str("username", username).Msg("employee not found")
		return "Employee not found", errors.New("employee not found")
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
