package app

import (
	"auth-service/internal/domain/entity"
	mock_repo "auth-service/internal/ports/mocks/repo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

// TestUpdateEmployee_Execute_Success tests success on update role
func TestUpdateEmployee_Execute_Success(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)

	updateEmployee := NewUpdateEmployee(mockEmployeeRepo)

	username := "john_doe"
	role := "admin"
	requester := "admin_user"

	employee := &entity.Employee{
		Username:  username,
		Role:      "viewer",
		Status:    entity.EmployeeStatusValid,
		UpdatedBy: "previous_user",
	}

	updatedEmployee := &entity.Employee{
		Username:  username,
		Role:      role, // Updated role
		Status:    entity.EmployeeStatusValid,
		UpdatedBy: requester,
	}

	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(employee, nil)
	mockEmployeeRepo.On("UpdateEmployee", mock.MatchedBy(func(emp *entity.Employee) bool {
		return emp.Username == username && emp.Role == role && emp.UpdatedBy == requester
	})).Return(updatedEmployee, nil)

	message, err := updateEmployee.Execute(username, role, requester)

	assert.NoError(t, err)
	assert.Equal(t, "Employee role updated successfully", message)
	mockEmployeeRepo.AssertExpectations(t)
}

// TestUpdateEmployee_Execute_SuccessDifferentRoles tests success for different roles
func TestUpdateEmployee_Execute_SuccessDifferentRoles(t *testing.T) {
	testCases := []struct {
		name string
		role string
	}{
		{"admin", "admin"},
		{"viewer", "viewer"},
		{"editor", "editor"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
			updateEmployee := NewUpdateEmployee(mockEmployeeRepo)

			username := "test_user"
			requester := "admin_user"

			employee := &entity.Employee{
				Username: username,
				Role:     "previous_role",
				Status:   entity.EmployeeStatusValid,
			}

			mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(employee, nil)
			mockEmployeeRepo.On("UpdateEmployee", mock.AnythingOfType("*entity.Employee")).Return(employee, nil)

			message, err := updateEmployee.Execute(username, tc.role, requester)

			assert.NoError(t, err)
			assert.Equal(t, "Employee role updated successfully", message)
			mockEmployeeRepo.AssertExpectations(t)
		})
	}
}
