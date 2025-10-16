package app

import (
	"auth-service/internal/common"
	"auth-service/internal/domain/entity"
	custom_err "auth-service/internal/domain/error"
	"auth-service/internal/logging"
	"auth-service/internal/observability/metrics"
	"auth-service/internal/ports"
	"regexp"
	"strings"
)

// CreateEmployee is a use-case for creating a new employee
type CreateEmployee struct {
	EmployeeRepo ports.EmployeeRepo
	Hashing      ports.Hashing
}

// NewCreateEmployee creates a new CreateEmployee use-case
func NewCreateEmployee(employeeRepo ports.EmployeeRepo, hashing ports.Hashing) *CreateEmployee {
	return &CreateEmployee{
		EmployeeRepo: employeeRepo,
		Hashing:      hashing,
	}
}

// Execute creates a new employee if they don't already exist
func (a *CreateEmployee) Execute(username, password, role, requester string) (string, error) {
	metrics.IncRequestActive()
	defer metrics.DecRequestActive()

	var err error
	defer func() {
		metrics.RecordOperation("create_employee", err)
	}()

	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	role = strings.TrimSpace(role)

	if username == "" || password == "" || role == "" {
		logging.Logger.Warn().Msg("Invalid request")
		return "Missing required data (username, password, role)", custom_err.ErrMissingRequiredData
	}

	if username == common.SystemUserUsername {
		logging.Logger.Warn().Err(custom_err.ErrInvalidUsername).Str("username", username).Msg("Invalid request")
		return "Username cannot be 'system'", custom_err.ErrInvalidUsername
	}

	re := regexp.MustCompile(`^[a-z]+(_[a-z]+)*$`)
	if !re.MatchString(username) {
		logging.Logger.Warn().Err(custom_err.ErrInvalidUsername).Str("username", username).Msg("Invalid username")
		return "Username supports only lowercase and '_' (in middle only)", custom_err.ErrInvalidUsername
	}

	existingEmployee, _ := a.EmployeeRepo.GetEmployeeByUsername(username)
	if existingEmployee != nil && existingEmployee.Status == entity.EmployeeStatusValid {
		logging.Logger.Warn().Err(custom_err.ErrEmployeeAlreadyExists).Str("username", username).Msg("Invalid request")
		return "Employee already exists", custom_err.ErrEmployeeAlreadyExists
	}

	if role != entity.EmployeeRoleAdmin && role != entity.EmployeeRoleViewer && role != entity.EmployeeRoleEditor {
		logging.Logger.Warn().Err(custom_err.ErrInvalidRole).Str("role", role).Msg("Invalid request")
		return "Invalid role", custom_err.ErrInvalidRole
	}

	hashedPassword, err := a.Hashing.HashData(password)
	if err != nil {
		logging.Logger.Warn().Err(err).Msg("unable to hash password")
		return "Failed to encrypt data", err
	}

	employee, err := entity.NewEmployee(
		username,
		hashedPassword,
		role,
		entity.EmployeeAuthMethodPassword,
		requester,
	)

	if err != nil {
		logging.Logger.Error().Err(err).Msg("unable to generate employee")
		return "Failed to create employee", err
	}

	_, err = a.EmployeeRepo.CreateEmployee(employee)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("unable to create employee")
		return "Failed to create employee", err
	}

	return "Employee created successfully", nil
}
