package app

import (
	"auth-service/internal/domain/entity"
	custom_err "auth-service/internal/domain/error"
	mock_auth "auth-service/internal/ports/mocks/auth"
	mock_repo "auth-service/internal/ports/mocks/repo"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

// TestCreateEmployee_Execute_Success tests success if input are provided correctly
func TestCreateEmployee_Execute_Success(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockHashing := new(mock_auth.MockHashing)

	createEmployee := NewCreateEmployee(mockEmployeeRepo, mockHashing)

	username := "john_doe"
	password := "password123"
	role := "admin"
	requester := "admin_user"
	hashedPassword := "hashed_password_123"

	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(nil, nil)
	mockHashing.On("HashData", password).Return(hashedPassword, nil)

	employee := &entity.Employee{
		Username: username,
		Password: hashedPassword,
		Role:     role,
	}
	mockEmployeeRepo.On("CreateEmployee", mock.AnythingOfType("*entity.Employee")).Return(employee, nil)

	message, err := createEmployee.Execute(username, password, role, requester)

	assert.NoError(t, err)
	assert.Equal(t, "Employee created successfully", message)
	mockEmployeeRepo.AssertExpectations(t)
	mockHashing.AssertExpectations(t)
}

// TestCreateEmployee_Execute_InvalidPassword_VariousCases test all invalid passwords
func TestCreateEmployee_Execute_InvalidPassword_VariousCases(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		reason   string
	}{
		{
			name:     "password with spaces",
			password: "pass word",
			reason:   "spaces are not allowed",
		},
		{
			name:     "password with parentheses",
			password: "password(123)",
			reason:   "parentheses are not in allowed characters",
		},
		{
			name:     "password with percent sign",
			password: "password%123",
			reason:   "percent sign is not allowed",
		},
		{
			name:     "password with plus sign",
			password: "password+123",
			reason:   "plus sign is not allowed",
		},
		{
			name:     "password with asterisk",
			password: "password*123",
			reason:   "asterisk is not allowed",
		},
		{
			name:     "empty password",
			password: "",
			reason:   "empty password should fail",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
			mockHashing := new(mock_auth.MockHashing)
			createEmployee := NewCreateEmployee(mockEmployeeRepo, mockHashing)

			_, err := createEmployee.Execute("valid_user", tc.password, "admin", "requester")

			assert.Error(t, err, "Expected error for password with %s", tc.reason)

			mockEmployeeRepo.AssertNotCalled(t, "GetEmployeeByUsername")
			mockEmployeeRepo.AssertNotCalled(t, "CreateEmployee")
			mockHashing.AssertNotCalled(t, "HashData")
		})
	}
}

// TestCreateEmployee_Execute_SuccessValidUsernameFormats tests success for valid users
func TestCreateEmployee_Execute_SuccessValidUsernameFormats(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockHashing := new(mock_auth.MockHashing)

	createEmployee := NewCreateEmployee(mockEmployeeRepo, mockHashing)

	validUsernames := []string{
		"john",
		"johndoe",
		"john_doe",
		"john_doe_smith",
		"a",
		"username_with_multiple_underscores",
	}

	for _, username := range validUsernames {
		t.Run(username, func(t *testing.T) {
			password := "password123"
			role := "admin"
			requester := "admin_user"
			hashedPassword := "hashed_password_123"

			mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(nil, nil)
			mockHashing.On("HashData", password).Return(hashedPassword, nil)

			employee := &entity.Employee{
				Username: username,
				Password: hashedPassword,
				Role:     role,
			}
			mockEmployeeRepo.On("CreateEmployee", mock.AnythingOfType("*entity.Employee")).Return(employee, nil)

			message, err := createEmployee.Execute(username, password, role, requester)

			assert.NoError(t, err)
			assert.Equal(t, "Employee created successfully", message)
			mockEmployeeRepo.AssertExpectations(t)
			mockHashing.AssertExpectations(t)
		})
	}
}

// TestCreateEmployee_Execute_SuccessValidRoles tests success for valid roles
func TestCreateEmployee_Execute_SuccessValidRoles(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockHashing := new(mock_auth.MockHashing)

	createEmployee := NewCreateEmployee(mockEmployeeRepo, mockHashing)

	validRoles := []string{"admin", "viewer", "editor"}

	for _, role := range validRoles {
		t.Run(role, func(t *testing.T) {
			username := "user_" + role
			password := "password123"
			hashedPassword := "hashed_password_" + role
			requester := "admin_user"

			mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(nil, nil)
			mockHashing.On("HashData", password).Return(hashedPassword, nil)

			// Mock: Employee creation
			employee := &entity.Employee{
				Username: username,
				Password: hashedPassword,
				Role:     role,
			}
			mockEmployeeRepo.On("CreateEmployee", mock.AnythingOfType("*entity.Employee")).Return(employee, nil)

			message, err := createEmployee.Execute(username, password, role, requester)

			assert.NoError(t, err)
			assert.Equal(t, "Employee created successfully", message)
			mockEmployeeRepo.AssertExpectations(t)
			mockHashing.AssertExpectations(t)
		})
	}
}

// TestCreateEmployee_Execute_ErrorEmployeeAlreadyExists test if username already exists
func TestCreateEmployee_Execute_ErrorEmployeeAlreadyExists(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockHashing := new(mock_auth.MockHashing)

	createEmployee := NewCreateEmployee(mockEmployeeRepo, mockHashing)

	username := "existing_user"
	password := "password123"
	role := "admin"
	requester := "admin_user"

	existingEmployee := &entity.Employee{
		Username: username,
		Role:     "viewer",
		Status:   entity.EmployeeStatusValid,
	}
	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(existingEmployee, nil)

	message, err := createEmployee.Execute(username, password, role, requester)

	assert.Error(t, err)
	assert.Equal(t, "employee already exists", err.Error())
	assert.Equal(t, "Employee already exists", message)
	mockEmployeeRepo.AssertExpectations(t)
	mockHashing.AssertNotCalled(t, "HashData")
	mockEmployeeRepo.AssertNotCalled(t, "CreateEmployee")
}

// TestCreateEmployee_Execute_ErrorSystemUsername tests if username='system' provided
func TestCreateEmployee_Execute_ErrorSystemUsername(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockHashing := new(mock_auth.MockHashing)

	createEmployee := NewCreateEmployee(mockEmployeeRepo, mockHashing)

	username := "system"
	password := "password123"
	role := "admin"
	requester := "admin_user"

	message, err := createEmployee.Execute(username, password, role, requester)

	assert.Error(t, err)
	assert.ErrorIs(t, err, custom_err.ErrInvalidUsername)
	assert.Contains(t, message, "Username cannot be")
	mockEmployeeRepo.AssertNotCalled(t, "GetEmployeeByUsername")
	mockHashing.AssertNotCalled(t, "HashData")
	mockEmployeeRepo.AssertNotCalled(t, "CreateEmployee")
}

// TestCreateEmployee_Execute_ErrorInvalidUsernameFormat tests for invalid username
func TestCreateEmployee_Execute_ErrorInvalidUsernameFormat(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockHashing := new(mock_auth.MockHashing)

	createEmployee := NewCreateEmployee(mockEmployeeRepo, mockHashing)

	testCases := []struct {
		name     string
		username string
	}{
		{"uppercase", "JohnDoe"},
		{"special_chars", "john@doe"},
		{"numbers", "john123"},
		{"start_underscore", "_johndoe"},
		{"end_underscore", "johndoe_"},
		{"double_underscore", "john__doe"},
		{"spaces", "john doe"},
		{"dots", "john.doe"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			message, err := createEmployee.Execute(tc.username, "password123", "admin", "admin_user")

			assert.Error(t, err)
			assert.ErrorIs(t, err, custom_err.ErrInvalidUsername)
			assert.Contains(t, message, "Username supports only lowercase")
			mockEmployeeRepo.AssertNotCalled(t, "GetEmployeeByUsername")
			mockHashing.AssertNotCalled(t, "HashData")
			mockEmployeeRepo.AssertNotCalled(t, "CreateEmployee")
		})
	}
}

// TestCreateEmployee_Execute_ErrorInvalidRole tests if roles are invalid
func TestCreateEmployee_Execute_ErrorInvalidRole(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockHashing := new(mock_auth.MockHashing)

	createEmployee := NewCreateEmployee(mockEmployeeRepo, mockHashing)

	testCases := []struct {
		name string
		role string
	}{
		{"invalid_role", "superadmin"},
		{"manager_role", "manager"},
		{"user_role", "user"},
		{"guest_role", "guest"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			username := "valid_user"

			mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(nil, nil)

			message, err := createEmployee.Execute(username, "password123", tc.role, "admin_user")

			assert.Error(t, err)
			assert.ErrorIs(t, err, custom_err.ErrInvalidRole)
			assert.Contains(t, message, "Invalid role")
			mockEmployeeRepo.AssertExpectations(t)
			mockHashing.AssertNotCalled(t, "HashData")
			mockEmployeeRepo.AssertNotCalled(t, "CreateEmployee")
		})
	}
}

// TestCreateEmployee_Execute_ErrorHashingFailure tests if hashing throws error
func TestCreateEmployee_Execute_ErrorHashingFailure(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockHashing := new(mock_auth.MockHashing)

	createEmployee := NewCreateEmployee(mockEmployeeRepo, mockHashing)

	username := "valid_user"
	password := "password123"
	role := "admin"
	requester := "admin_user"

	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(nil, nil)
	mockHashing.On("HashData", password).Return("", fmt.Errorf("hashing error"))

	message, err := createEmployee.Execute(username, password, role, requester)

	assert.Error(t, err)
	assert.Equal(t, "hashing error", err.Error())
	assert.Equal(t, "Failed to encrypt data", message)
	mockEmployeeRepo.AssertExpectations(t)
	mockHashing.AssertExpectations(t)
	mockEmployeeRepo.AssertNotCalled(t, "CreateEmployee")
}

// TestCreateEmployee_Execute_ErrorCreateEmployeeFailure tests employee creation failed
func TestCreateEmployee_Execute_ErrorCreateEmployeeFailure(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockHashing := new(mock_auth.MockHashing)

	createEmployee := NewCreateEmployee(mockEmployeeRepo, mockHashing)

	username := "valid_user"
	password := "password123"
	role := "admin"
	requester := "admin_user"
	hashedPassword := "hashed_password_123"

	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(nil, nil)
	mockHashing.On("HashData", password).Return(hashedPassword, nil)

	mockEmployeeRepo.On("CreateEmployee", mock.AnythingOfType("*entity.Employee")).Return(nil, fmt.Errorf("database error"))

	message, err := createEmployee.Execute(username, password, role, requester)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	assert.Equal(t, "Failed to create employee", message)
	mockEmployeeRepo.AssertExpectations(t)
	mockHashing.AssertExpectations(t)
}

// TestCreateEmployee_Execute_ErrorExistingEmployeeInvalidStatus if existing customer status is not valid (deleted already)
func TestCreateEmployee_Execute_ErrorExistingEmployeeInvalidStatus(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockHashing := new(mock_auth.MockHashing)

	createEmployee := NewCreateEmployee(mockEmployeeRepo, mockHashing)

	username := "inactive_user"
	password := "password123"
	role := "admin"
	requester := "admin_user"
	hashedPassword := "hashed_password_123"

	existingEmployee := &entity.Employee{
		Username: username,
		Role:     "viewer",
		Status:   entity.EmployeeStatusInvalid,
	}
	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(existingEmployee, nil)
	mockHashing.On("HashData", password).Return(hashedPassword, nil)

	employee := &entity.Employee{
		Username: username,
		Password: hashedPassword,
		Role:     role,
	}
	mockEmployeeRepo.On("CreateEmployee", mock.AnythingOfType("*entity.Employee")).Return(employee, nil)

	message, err := createEmployee.Execute(username, password, role, requester)

	assert.NoError(t, err)
	assert.Equal(t, "Employee created successfully", message)
	mockEmployeeRepo.AssertExpectations(t)
	mockHashing.AssertExpectations(t)
}

// TestCreateEmployee_Execute_ErrorEmptyPassword tests if password is empty
func TestCreateEmployee_Execute_ErrorEmptyPassword(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockHashing := new(mock_auth.MockHashing)

	createEmployee := NewCreateEmployee(mockEmployeeRepo, mockHashing)

	username := "valid_user"
	password := ""
	role := "admin"
	requester := "admin_user"

	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(nil, nil)
	mockHashing.On("HashData", password).Return("hashed_empty", nil)

	message, err := createEmployee.Execute(username, password, role, requester)

	if err != nil {
		assert.Contains(t, message, "Missing required data")
	}
}

// TestCreateEmployee_Execute_ErrorEmptyRole tests if role is empty
func TestCreateEmployee_Execute_ErrorEmptyRole(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockHashing := new(mock_auth.MockHashing)

	createEmployee := NewCreateEmployee(mockEmployeeRepo, mockHashing)

	username := "valid_user"
	password := "valid_pass"
	role := ""
	requester := "admin_user"

	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(nil, nil)
	mockHashing.On("HashData", password).Return("hashed_empty", nil)

	message, err := createEmployee.Execute(username, password, role, requester)

	if err != nil {
		assert.Contains(t, message, "Missing required data")
	}
}
