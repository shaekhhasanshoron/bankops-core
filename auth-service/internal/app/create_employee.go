package app

import (
	"auth-service/internal/auth"
	"auth-service/internal/common"
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
func (c *CreateEmployee) Execute(username, password, role string) (string, error) {
	existingEmployee, _ := c.EmployeeRepo.GetEmployeeByUsername(username)
	if existingEmployee != nil && existingEmployee.Status == common.EmployeeStatusValid {
		return "Employee already exists", errors.New("employee already exists")
	}

	re := regexp.MustCompile(`^[a-z]+(_[a-z]+)*$`)
	if !re.MatchString(username) {
		logging.Logger.Warn().Err(errors.New("Invalid username")).Msg("username: " + username)
		return "Invalid username: supports only lowercase and '_' (in middle only)", errors.New("Invalid username")
	}

	// Step 2: Hash the password before saving
	hashedPassword, err := auth.HashData(password)
	if err != nil {
		logging.Logger.Warn().Err(err).Msg("unable to hash password")
		return "Failed to encrypt data", err
	}

	if role == "" || (role != common.EmployeeRoleAdmin && role != common.EmployeeRoleViewer && role != common.EmployeeRoleEditor) {
		logging.Logger.Warn().Err(errors.New("invalid role")).Msg("role: " + role)
		return "Invalid role", errors.New("invalid role")
	}

	employee := ports.Employee{
		Username: username,
		Password: hashedPassword,
		Role:     role,
	}

	_, err = c.EmployeeRepo.CreateEmployee(&employee)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("unable to create employee")
		return "Failed to create employee", err
	}

	return "Employee created successfully", nil
}
