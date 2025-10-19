package app

import (
	"auth-service/internal/domain/entity"
	custom_err "auth-service/internal/domain/error"
	"auth-service/internal/logging"
	"auth-service/internal/messaging"
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
		err = custom_err.ErrMissingRequiredData
		return "Missing required data username and role", err
	}

	if role != entity.EmployeeRoleAdmin && role != entity.EmployeeRoleViewer && role != entity.EmployeeRoleEditor {
		logging.Logger.Warn().Err(custom_err.ErrInvalidRole).Str("role", role).Msg("Invalid request")
		err = custom_err.ErrInvalidRole
		return "Invalid role", err
	}

	employee, err := a.EmployeeRepo.GetEmployeeByUsername(username)
	if err != nil {
		logging.Logger.Warn().Err(err).Str("username", username).Msg("employee not found")
		err = errors.New("employee not found")
		return "Employee not found", err
	}

	employee.Role = role
	employee.UpdatedBy = requester

	_, err = a.EmployeeRepo.UpdateEmployee(employee)
	if err != nil {
		logging.Logger.Warn().Err(err).Str("username", username).Msg("failed to update employee role")
		return "Failed to update employee role", err
	}

	_ = messaging.GetService().PublishToDefaultTopic(messaging.Message{Content: employee.ToString(), Status: true, Type: messaging.MessageTypeEmployeeUpdated})
	return "Employee role updated successfully", nil
}
