package app

import (
	"auth-service/internal/auth"
	"auth-service/internal/common"
	"auth-service/internal/domain/entity"
	"auth-service/internal/logging"
	"auth-service/internal/ports"
	"errors"
	"regexp"
)

// CreateEmployee is a use-case for creating a new employee
type CreateEmployee struct {
	EmployeeRepo ports.EmployeeRepo
	TokenSigner  *auth.TokenSigner
}

// NewCreateEmployee creates a new CreateEmployee use-case
func NewCreateEmployee(employeeRepo ports.EmployeeRepo, tokenSigner *auth.TokenSigner) *CreateEmployee {
	return &CreateEmployee{
		EmployeeRepo: employeeRepo,
		TokenSigner:  tokenSigner,
	}
}

// Execute creates a new employee if they don't already exist
func (c *CreateEmployee) Execute(username, password, role, requester string) (string, error) {
	existingEmployee, _ := c.EmployeeRepo.GetEmployeeByUsername(username)
	if existingEmployee != nil && existingEmployee.Status == entity.EmployeeStatusValid {
		return "Employee already exists", errors.New("employee already exists")
	}

	if username == common.SystemUserUsername {
		logging.Logger.Warn().Err(errors.New("Invalid username")).Msg("username: " + username)
		return "Invalid username: username cannot be 'system'", errors.New("Invalid username")
	}

	re := regexp.MustCompile(`^[a-z]+(_[a-z]+)*$`)
	if !re.MatchString(username) {
		logging.Logger.Warn().Err(errors.New("Invalid username")).Msg("username: " + username)
		return "Invalid username: supports only lowercase and '_' (in middle only)", errors.New("Invalid username")
	}

	if role == "" || (role != entity.EmployeeRoleAdmin && role != entity.EmployeeRoleViewer && role != entity.EmployeeRoleEditor) {
		logging.Logger.Warn().Err(errors.New("invalid role")).Msg("role: " + role)
		return "Invalid role", errors.New("invalid role")
	}

	hashedPassword, err := auth.HashData(password)
	if err != nil {
		logging.Logger.Warn().Err(err).Msg("unable to hash password")
		return "Failed to encrypt data", err
	}

	employee := ports.Employee{
		Username:  username,
		Password:  hashedPassword,
		Role:      role,
		Requester: requester,
	}

	_, err = c.EmployeeRepo.CreateEmployee(&employee)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("unable to create employee")
		return "Failed to create employee", err
	}

	return "Employee created successfully", nil
}
