package app

import (
	"auth-service/internal/domain/entity"
	mock_repo "auth-service/internal/ports/mocks/repo"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestListEmployee_Execute_Success tests success if inputs are provided correctly
func TestListEmployee_Execute_Success(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)

	listEmployee := NewListEmployee(mockEmployeeRepo)

	page := 1
	pageSize := 10
	sortOrder := "asc"

	employees := []*entity.Employee{
		{ID: "emp-1", Username: "john_doe", Role: "admin", Status: entity.EmployeeStatusValid},
		{ID: "emp-2", Username: "jane_smith", Role: "viewer", Status: entity.EmployeeStatusValid},
	}
	totalCount := int64(2)

	mockEmployeeRepo.On("ListEmployee", page, pageSize, sortOrder).Return(employees, totalCount, nil)

	result, resultTotal, totalPages, message, err := listEmployee.Execute(page, pageSize, sortOrder)

	assert.NoError(t, err)
	assert.Equal(t, employees, result)
	assert.Equal(t, totalCount, resultTotal)
	assert.Equal(t, int64(1), totalPages) // 2 items / 10 per page = 1 page
	assert.Equal(t, "Employee List", message)
	mockEmployeeRepo.AssertExpectations(t)
}

// TestListEmployee_Execute_SuccessWithPagination tests success for valida pagination
func TestListEmployee_Execute_SuccessWithPagination(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)

	listEmployee := NewListEmployee(mockEmployeeRepo)

	page := 2
	pageSize := 5
	sortOrder := "desc"

	employees := []*entity.Employee{
		{ID: "emp-6", Username: "user6", Role: "editor"},
		{ID: "emp-7", Username: "user7", Role: "viewer"},
	}
	totalCount := int64(12)

	mockEmployeeRepo.On("ListEmployee", page, pageSize, sortOrder).Return(employees, totalCount, nil)

	result, resultTotal, totalPages, message, err := listEmployee.Execute(page, pageSize, sortOrder)

	assert.NoError(t, err)
	assert.Equal(t, employees, result)
	assert.Equal(t, totalCount, resultTotal)
	assert.Equal(t, int64(3), totalPages)
	assert.Equal(t, "Employee List", message)
	mockEmployeeRepo.AssertExpectations(t)
}

// TestListEmployee_Execute_SuccessEmptyResult tests success for empty result
func TestListEmployee_Execute_SuccessEmptyResult(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)

	listEmployee := NewListEmployee(mockEmployeeRepo)

	page := 1
	pageSize := 10
	sortOrder := "asc"

	employees := []*entity.Employee{}
	totalCount := int64(0)

	mockEmployeeRepo.On("ListEmployee", page, pageSize, sortOrder).Return(employees, totalCount, nil)

	result, resultTotal, totalPages, message, err := listEmployee.Execute(page, pageSize, sortOrder)

	assert.NoError(t, err)
	assert.Equal(t, employees, result)
	assert.Equal(t, totalCount, resultTotal)
	assert.Equal(t, int64(0), totalPages)
	assert.Equal(t, "Employee List", message)
	mockEmployeeRepo.AssertExpectations(t)
}

// TestListEmployee_Execute_PageBoundaryValues tests boundary page values
func TestListEmployee_Execute_PageBoundaryValues(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)

	listEmployee := NewListEmployee(mockEmployeeRepo)

	testCases := []struct {
		name         string
		page         int
		expectedPage int
	}{
		{"page_0", 0, 1},
		{"page_1", 1, 1},
		{"page_2", 2, 2},
		{"page_negative", -1, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pageSize := 10
			sortOrder := "asc"

			employees := []*entity.Employee{}
			totalCount := int64(0)

			mockEmployeeRepo.On("ListEmployee", tc.expectedPage, pageSize, sortOrder).Return(employees, totalCount, nil)

			result, resultTotal, totalPages, message, err := listEmployee.Execute(tc.page, pageSize, sortOrder)

			assert.NoError(t, err)
			assert.Equal(t, employees, result)
			assert.Equal(t, totalCount, resultTotal)
			assert.Equal(t, int64(0), totalPages)
			assert.Equal(t, "Employee List", message)
			mockEmployeeRepo.AssertExpectations(t)
		})
	}
}

// TestListEmployee_Execute_PageSizeBoundaryValues tests boundary of pageSize values
func TestListEmployee_Execute_PageSizeBoundaryValues(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)

	listEmployee := NewListEmployee(mockEmployeeRepo)

	testCases := []struct {
		name         string
		pageSize     int
		expectedSize int
	}{
		{"page_size_0", 0, 50},
		{"page_size_1", 1, 1},
		{"page_size_25", 25, 25},
		{"page_size_50", 50, 50},
		{"page_size_51", 51, 50},
		{"page_size_100", 100, 50},
		{"page_size_negative", -5, 50},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			page := 1
			sortOrder := "asc"

			employees := []*entity.Employee{}
			totalCount := int64(0)

			mockEmployeeRepo.On("ListEmployee", page, tc.expectedSize, sortOrder).Return(employees, totalCount, nil)

			result, resultTotal, totalPages, message, err := listEmployee.Execute(page, tc.pageSize, sortOrder)

			assert.NoError(t, err)
			assert.Equal(t, employees, result)
			assert.Equal(t, totalCount, resultTotal)
			assert.Equal(t, int64(0), totalPages)
			assert.Equal(t, "Employee List", message)
			mockEmployeeRepo.AssertExpectations(t)
		})
	}
}

// TestListEmployee_Execute_TotalPagesCalculation tests if totalPage calculation is correctly working
func TestListEmployee_Execute_TotalPagesCalculation(t *testing.T) {
	testCases := []struct {
		name          string
		totalCount    int64
		pageSize      int
		expectedPages int64
	}{
		{"exact_division", 100, 10, 10},
		{"round_up_division", 101, 10, 11},
		{"single_page", 5, 10, 1},
		{"empty_result", 0, 10, 0},
		{"one_item", 1, 10, 1},
		{"page_size_1", 10, 1, 10},
		{"large_numbers", 1000, 50, 20},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
			listEmployee := NewListEmployee(mockEmployeeRepo)

			page := 1
			sortOrder := "asc"

			employees := []*entity.Employee{}
			if tc.totalCount > 0 {
				employees = []*entity.Employee{{ID: "emp-1", Username: "test"}}
			}

			mockEmployeeRepo.On("ListEmployee", page, tc.pageSize, sortOrder).Return(employees, tc.totalCount, nil)

			_, resultTotal, totalPages, message, err := listEmployee.Execute(page, tc.pageSize, sortOrder)

			assert.NoError(t, err)
			assert.Equal(t, tc.totalCount, resultTotal, "Total count mismatch for test case: %s", tc.name)
			assert.Equal(t, tc.expectedPages, totalPages, "Total pages mismatch for test case: %s", tc.name)
			assert.Equal(t, "Employee List", message)
			mockEmployeeRepo.AssertExpectations(t)
		})
	}
}

// TestListEmployee_Execute_SuccessWithSortOrders tests success with different order
func TestListEmployee_Execute_SuccessWithSortOrders(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)

	listEmployee := NewListEmployee(mockEmployeeRepo)

	testCases := []struct {
		name      string
		sortOrder string
	}{
		{"ascending", "asc"},
		{"descending", "desc"},
		{"empty", ""},
		{"uppercase_asc", "ASC"},
		{"uppercase_desc", "DESC"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			page := 1
			pageSize := 10

			employees := []*entity.Employee{
				{ID: "emp-1", Username: "user1", Role: "admin"},
			}
			totalCount := int64(1)

			mockEmployeeRepo.On("ListEmployee", page, pageSize, tc.sortOrder).Return(employees, totalCount, nil)

			result, resultTotal, totalPages, message, err := listEmployee.Execute(page, pageSize, tc.sortOrder)

			assert.NoError(t, err)
			assert.Equal(t, employees, result)
			assert.Equal(t, totalCount, resultTotal)
			assert.Equal(t, int64(1), totalPages)
			assert.Equal(t, "Employee List", message)
			mockEmployeeRepo.AssertExpectations(t)
		})
	}
}

// TestListEmployee_Execute_SuccessLargeDataset tests for large employee list
func TestListEmployee_Execute_SuccessLargeDataset(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)

	listEmployee := NewListEmployee(mockEmployeeRepo)

	page := 10
	pageSize := 50
	sortOrder := "desc"

	employees := make([]*entity.Employee, 50)
	for i := 0; i < 50; i++ {
		employees[i] = &entity.Employee{
			ID:       fmt.Sprintf("emp-%d", (page-1)*pageSize+i+1),
			Username: fmt.Sprintf("user%d", (page-1)*pageSize+i+1),
			Role:     "viewer",
		}
	}
	totalCount := int64(10000)

	mockEmployeeRepo.On("ListEmployee", page, pageSize, sortOrder).Return(employees, totalCount, nil)

	result, resultTotal, totalPages, message, err := listEmployee.Execute(page, pageSize, sortOrder)

	assert.NoError(t, err)
	assert.Equal(t, employees, result)
	assert.Equal(t, totalCount, resultTotal)
	assert.Equal(t, int64(200), totalPages) // 10000 items / 50 per page = 200 pages
	assert.Equal(t, "Employee List", message)
	mockEmployeeRepo.AssertExpectations(t)
}
