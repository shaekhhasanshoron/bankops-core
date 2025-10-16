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

// TestAuthenticate_Execute_Success tests success authentication for correct input
func TestAuthenticate_Execute_Success(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockTokenSigner := new(mock_auth.MockTokenSigner)
	mockHashing := new(mock_auth.MockHashing)

	authenticate := NewAuthenticate(mockEmployeeRepo, mockTokenSigner, mockHashing)

	username := "testuser"
	password := "password123"
	hashedPassword := "hashed_password_123"

	employee := &entity.Employee{
		Username: username,
		Password: hashedPassword,
		Role:     "admin",
		Status:   entity.EmployeeStatusValid,
	}
	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(employee, nil)
	mockHashing.On("HashData", password).Return(hashedPassword, nil)

	expectedToken := "jwt-token-123"
	expectedRefreshToken := "refresh-token-456"
	mockTokenSigner.On("SignJWT", username, employee.Role, mock.Anything, mock.Anything).Return(expectedToken, nil)
	mockTokenSigner.On("SignJWTRefreshToken", username, mock.Anything).Return(expectedRefreshToken, nil)

	token, refreshToken, err := authenticate.Execute(username, password)

	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
	assert.Equal(t, expectedRefreshToken, refreshToken)
	mockEmployeeRepo.AssertExpectations(t)
	mockTokenSigner.AssertExpectations(t)
	mockHashing.AssertExpectations(t)
}

// TestAuthenticate_Execute_ErrorUserNotFound tests if user not found
func TestAuthenticate_Execute_ErrorUserNotFound(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockTokenSigner := new(mock_auth.MockTokenSigner)
	mockHashing := new(mock_auth.MockHashing)

	authenticate := NewAuthenticate(mockEmployeeRepo, mockTokenSigner, mockHashing)

	username := "nonexistent"
	password := "password123"

	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(nil, fmt.Errorf("user not found"))

	token, refreshToken, err := authenticate.Execute(username, password)

	assert.Error(t, err)
	assert.Equal(t, "invalid credentials or user not found", err.Error())
	assert.Empty(t, token)
	assert.Empty(t, refreshToken)
	mockEmployeeRepo.AssertExpectations(t)
	mockTokenSigner.AssertNotCalled(t, "SignJWT")
	mockHashing.AssertNotCalled(t, "HashData")
}

// TestAuthenticate_Execute_ErrorEmptyUsername tests if empty username is provided
func TestAuthenticate_Execute_ErrorEmptyUsername(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockTokenSigner := new(mock_auth.MockTokenSigner)
	mockHashing := new(mock_auth.MockHashing)

	authenticate := NewAuthenticate(mockEmployeeRepo, mockTokenSigner, mockHashing)

	token, refreshToken, err := authenticate.Execute("", "password123")

	assert.Error(t, err)
	assert.ErrorIs(t, err, custom_err.ErrInvalidUsername)
	assert.Empty(t, token)
	assert.Empty(t, refreshToken)
	mockEmployeeRepo.AssertNotCalled(t, "GetEmployeeByUsername")
	mockTokenSigner.AssertNotCalled(t, "SignJWT")
	mockHashing.AssertNotCalled(t, "HashData")
}

// TestAuthenticate_Execute_ErrorDatabaseFailure tests if database throws error
func TestAuthenticate_Execute_ErrorDatabaseFailure(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockTokenSigner := new(mock_auth.MockTokenSigner)
	mockHashing := new(mock_auth.MockHashing)

	authenticate := NewAuthenticate(mockEmployeeRepo, mockTokenSigner, mockHashing)

	username := "testuser"
	password := "password123"

	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(nil, fmt.Errorf("database connection failed"))

	token, refreshToken, err := authenticate.Execute(username, password)

	assert.Error(t, err)
	assert.Equal(t, "invalid credentials or user not found", err.Error())
	assert.Empty(t, token)
	assert.Empty(t, refreshToken)
	mockEmployeeRepo.AssertExpectations(t)
	mockTokenSigner.AssertNotCalled(t, "SignJWT")
	mockHashing.AssertNotCalled(t, "HashData")
}

// TestAuthenticate_Execute_ErrorInvalidPassword if password is inavlid
func TestAuthenticate_Execute_ErrorInvalidPassword(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockTokenSigner := new(mock_auth.MockTokenSigner)
	mockHashing := new(mock_auth.MockHashing)

	authenticate := NewAuthenticate(mockEmployeeRepo, mockTokenSigner, mockHashing)

	username := "testuser"
	password := "wrongpassword"
	storedHashedPassword := "stored_hashed_password"
	inputHashedPassword := "input_hashed_password" // Different from stored

	employee := &entity.Employee{
		Username: username,
		Password: storedHashedPassword,
		Role:     "admin",
		Status:   entity.EmployeeStatusValid,
	}
	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(employee, nil)
	mockHashing.On("HashData", password).Return(inputHashedPassword, nil)

	token, refreshToken, err := authenticate.Execute(username, password)

	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
	assert.Empty(t, token)
	assert.Empty(t, refreshToken)
	mockEmployeeRepo.AssertExpectations(t)
	mockTokenSigner.AssertNotCalled(t, "SignJWT")
	mockHashing.AssertExpectations(t)
}

// TestAuthenticate_Execute_ErrorUserInvalidStatus tests user is invalid
func TestAuthenticate_Execute_ErrorUserInvalidStatus(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockTokenSigner := new(mock_auth.MockTokenSigner)
	mockHashing := new(mock_auth.MockHashing)

	authenticate := NewAuthenticate(mockEmployeeRepo, mockTokenSigner, mockHashing)

	username := "testuser"
	password := "password123"
	hashedPassword := "hashed_password_123"

	employee := &entity.Employee{
		Username: username,
		Password: hashedPassword,
		Role:     "admin",
		Status:   entity.EmployeeActiveStatusDeactivated,
	}
	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(employee, nil)
	mockHashing.On("HashData", password).Return(hashedPassword, nil)

	token, refreshToken, err := authenticate.Execute(username, password)

	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
	assert.Empty(t, token)
	assert.Empty(t, refreshToken)
	mockEmployeeRepo.AssertExpectations(t)
	mockTokenSigner.AssertNotCalled(t, "SignJWT")
	mockHashing.AssertExpectations(t)
}

// TestAuthenticate_Execute_ErrorUserInvalidStatus tests hashing failure
func TestAuthenticate_Execute_ErrorHashingFailure(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockTokenSigner := new(mock_auth.MockTokenSigner)
	mockHashing := new(mock_auth.MockHashing)

	authenticate := NewAuthenticate(mockEmployeeRepo, mockTokenSigner, mockHashing)

	username := "testuser"
	password := "password123"

	employee := &entity.Employee{
		Username: username,
		Password: "hashed_password",
		Role:     "admin",
		Status:   entity.EmployeeStatusValid,
	}
	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(employee, nil)
	mockHashing.On("HashData", password).Return("", fmt.Errorf("hashing error"))

	token, refreshToken, err := authenticate.Execute(username, password)

	assert.Error(t, err)
	assert.Equal(t, "hashing error", err.Error())
	assert.Empty(t, token)
	assert.Empty(t, refreshToken)
	mockEmployeeRepo.AssertExpectations(t)
	mockTokenSigner.AssertNotCalled(t, "SignJWT")
	mockHashing.AssertExpectations(t)
}

// TestAuthenticate_Execute_ErrorJWTSigningFailure test jwt signing failure
func TestAuthenticate_Execute_ErrorJWTSigningFailure(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockTokenSigner := new(mock_auth.MockTokenSigner)
	mockHashing := new(mock_auth.MockHashing)

	authenticate := NewAuthenticate(mockEmployeeRepo, mockTokenSigner, mockHashing)

	username := "testuser"
	password := "password123"
	hashedPassword := "hashed_password_123"

	employee := &entity.Employee{
		Username: username,
		Password: hashedPassword,
		Role:     "admin",
		Status:   entity.EmployeeStatusValid,
	}
	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(employee, nil)
	mockHashing.On("HashData", password).Return(hashedPassword, nil)

	// Mock JWT signing to fail
	mockTokenSigner.On("SignJWT", username, employee.Role, mock.Anything, mock.Anything).Return("", fmt.Errorf("jwt signing error"))

	token, refreshToken, err := authenticate.Execute(username, password)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate token")
	assert.Empty(t, token)
	assert.Empty(t, refreshToken)
	mockEmployeeRepo.AssertExpectations(t)
	mockTokenSigner.AssertExpectations(t)
	mockHashing.AssertExpectations(t)
	mockTokenSigner.AssertNotCalled(t, "SignJWTRefreshToken")
}

// TestAuthenticate_Execute_ErrorRefreshTokenSigningFailure test refresh token signing failure
func TestAuthenticate_Execute_ErrorRefreshTokenSigningFailure(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockTokenSigner := new(mock_auth.MockTokenSigner)
	mockHashing := new(mock_auth.MockHashing)

	authenticate := NewAuthenticate(mockEmployeeRepo, mockTokenSigner, mockHashing)

	username := "testuser"
	password := "password123"
	hashedPassword := "hashed_password_123"

	employee := &entity.Employee{
		Username: username,
		Password: hashedPassword,
		Role:     "admin",
		Status:   entity.EmployeeStatusValid,
	}
	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(employee, nil)
	mockHashing.On("HashData", password).Return(hashedPassword, nil)

	// Mock JWT signing success but refresh token failure
	expectedToken := "jwt-token-123"
	mockTokenSigner.On("SignJWT", username, employee.Role, mock.Anything, mock.Anything).Return(expectedToken, nil)
	mockTokenSigner.On("SignJWTRefreshToken", username, mock.Anything).Return("", fmt.Errorf("refresh token error"))

	token, refreshToken, err := authenticate.Execute(username, password)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate refresh token")
	assert.Empty(t, token) // Even though JWT was generated, we return error so both should be empty
	assert.Empty(t, refreshToken)
	mockEmployeeRepo.AssertExpectations(t)
	mockTokenSigner.AssertExpectations(t)
	mockHashing.AssertExpectations(t)
}

// TestAuthenticate_Execute_SuccessDifferentRoles tests invalid roles
func TestAuthenticate_Execute_SuccessDifferentRoles(t *testing.T) {
	roles := []string{"admin", "editor", "viewer"}

	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
			mockTokenSigner := new(mock_auth.MockTokenSigner)
			mockHashing := new(mock_auth.MockHashing)

			authenticate := NewAuthenticate(mockEmployeeRepo, mockTokenSigner, mockHashing)

			username := "testuser"
			password := "password123"
			hashedPassword := "hashed_password_123"

			employee := &entity.Employee{
				Username: username,
				Password: hashedPassword,
				Role:     role,
				Status:   entity.EmployeeStatusValid,
			}
			mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(employee, nil)
			mockHashing.On("HashData", password).Return(hashedPassword, nil)

			expectedToken := "jwt-token-" + role
			expectedRefreshToken := "refresh-token-" + role
			mockTokenSigner.On("SignJWT", username, role, mock.Anything, mock.Anything).Return(expectedToken, nil)
			mockTokenSigner.On("SignJWTRefreshToken", username, mock.Anything).Return(expectedRefreshToken, nil)

			token, refreshToken, err := authenticate.Execute(username, password)

			assert.NoError(t, err)
			assert.Equal(t, expectedToken, token)
			assert.Equal(t, expectedRefreshToken, refreshToken)
			mockEmployeeRepo.AssertExpectations(t)
			mockTokenSigner.AssertExpectations(t)
			mockHashing.AssertExpectations(t)
		})
	}
}

// TestAuthenticate_Execute_ErrorEmptyPassword test if password is empty
func TestAuthenticate_Execute_ErrorEmptyPassword(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockTokenSigner := new(mock_auth.MockTokenSigner)
	mockHashing := new(mock_auth.MockHashing)

	authenticate := NewAuthenticate(mockEmployeeRepo, mockTokenSigner, mockHashing)

	username := "testuser"
	emptyPasswordHash := "hashed_empty_string"

	employee := &entity.Employee{
		Username: username,
		Password: "different_hashed_password", // Different from empty password hash
		Role:     "admin",
		Status:   entity.EmployeeStatusValid,
	}
	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(employee, nil)
	mockHashing.On("HashData", "").Return(emptyPasswordHash, nil)

	token, refreshToken, err := authenticate.Execute(username, "")

	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
	assert.Empty(t, token)
	assert.Empty(t, refreshToken)
	mockEmployeeRepo.AssertExpectations(t)
	mockTokenSigner.AssertNotCalled(t, "SignJWT")
	mockHashing.AssertExpectations(t)
}

// TestAuthenticate_Execute_SuccessWithSpecialCharacters test special characters
func TestAuthenticate_Execute_SuccessWithSpecialCharacters(t *testing.T) {
	mockEmployeeRepo := new(mock_repo.MockEmployeeRepo)
	mockTokenSigner := new(mock_auth.MockTokenSigner)
	mockHashing := new(mock_auth.MockHashing)

	authenticate := NewAuthenticate(mockEmployeeRepo, mockTokenSigner, mockHashing)

	username := "user@domain.com"
	password := "p@ssw0rd!@#$%^&*()"
	hashedPassword := "hashed_complex_password"

	employee := &entity.Employee{
		Username: username,
		Password: hashedPassword,
		Role:     "admin",
		Status:   entity.EmployeeStatusValid,
	}
	mockEmployeeRepo.On("GetEmployeeByUsername", username).Return(employee, nil)
	mockHashing.On("HashData", password).Return(hashedPassword, nil)

	expectedToken := "jwt-token-complex"
	expectedRefreshToken := "refresh-token-complex"
	mockTokenSigner.On("SignJWT", username, employee.Role, mock.Anything, mock.Anything).Return(expectedToken, nil)
	mockTokenSigner.On("SignJWTRefreshToken", username, mock.Anything).Return(expectedRefreshToken, nil)

	token, refreshToken, err := authenticate.Execute(username, password)

	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
	assert.Equal(t, expectedRefreshToken, refreshToken)
	mockEmployeeRepo.AssertExpectations(t)
	mockTokenSigner.AssertExpectations(t)
	mockHashing.AssertExpectations(t)
}
