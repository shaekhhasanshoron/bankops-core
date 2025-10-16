package app

import (
	"auth-service/internal/domain/entity"
	custom_err "auth-service/internal/domain/error"
	mock_repo "auth-service/internal/ports/mocks/repo"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
)

// TestDeleteEmployee_Execute_Success tests success for employee delete
func TestDeleteEmployee_Execute_Success(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)

	deleteEmployee := NewDeleteEmployee(mockEmployeeRepo)

	username := "john_doe"
	requester := "admin_user"

	employee := &entity.Employee{
		Username: username,
		Role:     "admin",
		Status:   entity.EmployeeStatusValid,
	}

	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(employee, nil)
	mockEmployeeRepo.On("DeleteEmployee", username, requester).Return(nil)

	message, err := deleteEmployee.Execute(username, requester)

	assert.NoError(t, err)
	assert.Equal(t, "Employee deleted successfully", message)
	mockEmployeeRepo.AssertExpectations(t)
}

// TestDeleteEmployee_Execute_SuccessWithDifferentUsernames tests success with valid usernames
func TestDeleteEmployee_Execute_SuccessWithDifferentUsernames(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)

	deleteEmployee := NewDeleteEmployee(mockEmployeeRepo)

	testCases := []struct {
		name     string
		username string
	}{
		{"simple", "johndoe"},
		{"with_underscore", "john_doe"},
		{"multiple_underscores", "john_doe_smith"},
		{"single_char", "a"},
		{"max_length", "very_long_username_that_is_valid"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requester := "admin_user"

			employee := &entity.Employee{
				Username: tc.username,
				Role:     "viewer",
				Status:   entity.EmployeeStatusValid,
			}

			mockEmployeeRepo.On("GetEmployeeByUsername", tc.username).Return(employee, nil)
			mockEmployeeRepo.On("DeleteEmployee", tc.username, requester).Return(nil)

			message, err := deleteEmployee.Execute(tc.username, requester)

			assert.NoError(t, err)
			assert.Equal(t, "Employee deleted successfully", message)
			mockEmployeeRepo.AssertExpectations(t)
		})
	}
}

// TestDeleteEmployee_Execute_SuccessWithDifferentEmployeeStatuses tests success for different status employee
func TestDeleteEmployee_Execute_SuccessWithDifferentEmployeeStatuses(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)

	deleteEmployee := NewDeleteEmployee(mockEmployeeRepo)

	testCases := []struct {
		name   string
		status string
	}{
		{"valid", entity.EmployeeStatusValid},
		{"invalid", entity.EmployeeStatusInvalid},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			username := "employee_user"
			requester := "admin_user"

			employee := &entity.Employee{
				Username: username,
				Role:     "admin",
				Status:   tc.status,
			}

			mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(employee, nil)
			mockEmployeeRepo.On("DeleteEmployee", username, requester).Return(nil)

			message, err := deleteEmployee.Execute(username, requester)

			assert.NoError(t, err)
			assert.Equal(t, "Employee deleted successfully", message)
			mockEmployeeRepo.AssertExpectations(t)
		})
	}
}

// TestDeleteEmployee_Execute_ErrorEmptyUsername tests if username is empty
func TestDeleteEmployee_Execute_ErrorEmptyUsername(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)

	deleteEmployee := NewDeleteEmployee(mockEmployeeRepo)

	testCases := []struct {
		name     string
		username string
	}{
		{"empty_string", ""},
		{"only_spaces", "   "},
		{"tabs_only", "\t\t"},
		{"spaces_and_tabs", "  \t  "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			message, err := deleteEmployee.Execute(tc.username, "admin_user")

			assert.Error(t, err)
			assert.Equal(t, custom_err.ErrMissingRequiredData, err)
			assert.Equal(t, "Missing required data (username)", message)
			mockEmployeeRepo.AssertNotCalled(t, "GetEmployeeByUsername")
			mockEmployeeRepo.AssertNotCalled(t, "DeleteEmployee")
		})
	}
}

// TestDeleteEmployee_Execute_ErrorEmployeeNotFound tests if employee not found
func TestDeleteEmployee_Execute_ErrorEmployeeNotFound(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)

	deleteEmployee := NewDeleteEmployee(mockEmployeeRepo)

	username := "nonexistent_user"
	requester := "admin_user"

	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(nil, nil)

	message, err := deleteEmployee.Execute(username, requester)

	assert.Error(t, err)
	assert.Equal(t, custom_err.ErrEmployeeNotFound, err)
	assert.Equal(t, "Employee not found", message)
	mockEmployeeRepo.AssertExpectations(t)
	mockEmployeeRepo.AssertNotCalled(t, "DeleteEmployee")
}

// TestDeleteEmployee_Execute_ErrorEmployeeNotFoundWithGormError test if database throws error
func TestDeleteEmployee_Execute_ErrorEmployeeNotFoundWithGormError(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)

	deleteEmployee := NewDeleteEmployee(mockEmployeeRepo)

	username := "nonexistent_user"
	requester := "admin_user"

	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(nil, gorm.ErrRecordNotFound)

	message, err := deleteEmployee.Execute(username, requester)

	assert.Error(t, err)
	assert.Equal(t, custom_err.ErrEmployeeNotFound, err)
	assert.Equal(t, "Employee not found", message)
	mockEmployeeRepo.AssertExpectations(t)
	mockEmployeeRepo.AssertNotCalled(t, "DeleteEmployee")
}

// TestDeleteEmployee_Execute_ErrorDatabaseGetEmployee tests if database throws error
func TestDeleteEmployee_Execute_ErrorDatabaseGetEmployee(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)

	deleteEmployee := NewDeleteEmployee(mockEmployeeRepo)

	username := "test_user"
	requester := "admin_user"

	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(nil, fmt.Errorf("database connection failed"))

	message, err := deleteEmployee.Execute(username, requester)

	assert.Error(t, err)
	assert.Equal(t, custom_err.ErrDatabase, err)
	assert.Equal(t, "Failed to get employee", message)
	mockEmployeeRepo.AssertExpectations(t)
	mockEmployeeRepo.AssertNotCalled(t, "DeleteEmployee")
}

// TestDeleteEmployee_Execute_ErrorDeleteEmployeeFailure tests if employee deletion fails
func TestDeleteEmployee_Execute_ErrorDeleteEmployeeFailure(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)

	deleteEmployee := NewDeleteEmployee(mockEmployeeRepo)

	username := "john_doe"
	requester := "admin_user"

	employee := &entity.Employee{
		Username: username,
		Role:     "admin",
		Status:   entity.EmployeeStatusValid,
	}

	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(employee, nil)
	mockEmployeeRepo.On("DeleteEmployee", username, requester).Return(fmt.Errorf("delete failed"))

	message, err := deleteEmployee.Execute(username, requester)

	assert.Error(t, err)
	assert.Equal(t, "delete failed", err.Error())
	assert.Equal(t, "Failed to delete employee", message)
	mockEmployeeRepo.AssertExpectations(t)
}

// TestDeleteEmployee_Execute_SuccessWithDifferentRequesters test success with different requesters
func TestDeleteEmployee_Execute_SuccessWithDifferentRequesters(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)

	deleteEmployee := NewDeleteEmployee(mockEmployeeRepo)

	testCases := []struct {
		name      string
		requester string
	}{
		{"admin_user", "admin_user"},
		{"system_user", "system"},
		{"manager_user", "manager_user"},
		{"user_with_underscore", "user_admin"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			username := "employee_to_delete"

			employee := &entity.Employee{
				Username: username,
				Role:     "editor",
				Status:   entity.EmployeeStatusValid,
			}

			mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(employee, nil)
			mockEmployeeRepo.On("DeleteEmployee", username, tc.requester).Return(nil)

			message, err := deleteEmployee.Execute(username, tc.requester)

			assert.NoError(t, err)
			assert.Equal(t, "Employee deleted successfully", message)
			mockEmployeeRepo.AssertExpectations(t)
		})
	}
}

// TestDeleteEmployee_Execute_ErrorSelfDeletion tests if your tries to delete own account
func TestDeleteEmployee_Execute_ErrorSelfDeletion(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)

	deleteEmployee := NewDeleteEmployee(mockEmployeeRepo)

	username := "admin_user"
	requester := "admin_user"

	message, err := deleteEmployee.Execute(username, requester)

	assert.Error(t, err)
	assert.ErrorIs(t, err, custom_err.ErrInvalidRequest)
	assert.Contains(t, message, "Cannot delete yourself")
	mockEmployeeRepo.AssertExpectations(t)
}

// TestDeleteEmployee_Execute_ErrorConcurrentDeletion tests concurrent deletion
func TestDeleteEmployee_Execute_ErrorConcurrentDeletion(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)

	deleteEmployee := NewDeleteEmployee(mockEmployeeRepo)

	username := "john_doe"
	requester := "admin_user"

	employee := &entity.Employee{
		Username: username,
		Role:     "admin",
		Status:   entity.EmployeeStatusValid,
	}

	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(employee, nil)

	mockEmployeeRepo.On("DeleteEmployee", username, requester).Return(fmt.Errorf("employee already being deleted"))

	message, err := deleteEmployee.Execute(username, requester)

	assert.Error(t, err)
	assert.Equal(t, "employee already being deleted", err.Error())
	assert.Equal(t, "Failed to delete employee", message)
	mockEmployeeRepo.AssertExpectations(t)
}

// TestDeleteEmployee_Execute_ErrorEmployeeAlreadyDeleted tests if employee already deleted
func TestDeleteEmployee_Execute_ErrorEmployeeAlreadyDeleted(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)

	deleteEmployee := NewDeleteEmployee(mockEmployeeRepo)

	username := "already_deleted"
	requester := "admin_user"

	employee := &entity.Employee{
		Username: username,
		Role:     "admin",
		Status:   entity.EmployeeStatusInvalid,
	}

	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(employee, nil)

	mockEmployeeRepo.On("DeleteEmployee", username, requester).Return(nil)

	message, err := deleteEmployee.Execute(username, requester)

	assert.NoError(t, err)
	assert.Equal(t, "Employee deleted successfully", message)
	mockEmployeeRepo.AssertExpectations(t)
}
