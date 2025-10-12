package app

import (
	"auth-service/internal/auth"
	"auth-service/internal/common"
	"auth-service/internal/logging"
	"auth-service/internal/ports"
	"errors"
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
		return "", errors.New("employee already exists")
	}

	// Step 2: Hash the password before saving
	hashedPassword, err := auth.HashData(password)
	if err != nil {
		logging.Logger.Warn().Err(err).Msg("unable to hash password")
		return "", err
	}

	employee := ports.Employee{
		Username: username,
		Password: hashedPassword,
		Role:     role,
	}

	_, err = c.EmployeeRepo.CreateEmployee(&employee)
	if err != nil {
		return "", err
	}

	return "Employee created successfully", nil
}
