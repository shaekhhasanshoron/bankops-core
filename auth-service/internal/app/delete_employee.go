package app

import (
	custom_err "auth-service/internal/domain/error"
	"auth-service/internal/logging"
	"auth-service/internal/messaging"
	"auth-service/internal/observability/metrics"
	"auth-service/internal/ports"
	"errors"
	"gorm.io/gorm"
	"strings"
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
	metrics.IncRequestActive()
	defer metrics.DecRequestActive()

	var err error
	defer func() {
		metrics.RecordOperation("delete_employee", err)
	}()

	username = strings.TrimSpace(username)
	requester = strings.TrimSpace(requester)

	if username == "" {
		logging.Logger.Err(custom_err.ErrMissingRequiredData).Msg("Invalid request")
		return "Missing required data (username)", custom_err.ErrMissingRequiredData
	}

	if username == requester {
		logging.Logger.Err(custom_err.ErrInvalidRequest).Msg("Invalid request: Cannot delete own-self")
		return "Cannot delete yourself", custom_err.ErrInvalidRequest
	}

	employee, err := a.EmployeeRepo.GetEmployeeByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logging.Logger.Warn().Err(err).Str("username", username).Msg("Employee not found")
			err = custom_err.ErrEmployeeNotFound
			return "Employee not found", err
		}

		logging.Logger.Error().Err(err).Str("username", username).Msg("Failed to get employee")
		return "Failed to get employee", custom_err.ErrDatabase
	}

	if employee == nil {
		err = custom_err.ErrEmployeeNotFound
		logging.Logger.Warn().Err(err).Str("username", username).Msg("Employee not found")
		return "Employee not found", err
	}

	err = a.EmployeeRepo.DeleteEmployee(username, requester)
	if err != nil {
		logging.Logger.Warn().Err(err).Str("username", username).Msg("failed to mark employee as invalid")
		return "Failed to delete employee", err
	}

	_ = messaging.GetService().PublishToDefaultTopic(messaging.Message{Content: username, Status: true, Type: messaging.MessageTypeEmployeeDeleted})
	return "Employee deleted successfully", nil
}
