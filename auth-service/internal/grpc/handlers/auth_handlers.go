package handlers

import (
	"auth-service/api/protogen/authservice/proto"
	"auth-service/internal/app"
	"auth-service/internal/logging"
	"context"
	"fmt"
)

// AuthHandler implements the AuthServiceServer interface.
type AuthHandler struct {
	proto.UnimplementedAuthServiceServer
	authenticate   *app.Authenticate
	createEmployee *app.CreateEmployee
	updateEmployee *app.UpdateEmployee
	deleteEmployee *app.DeleteEmployee
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authenticate *app.Authenticate, createEmployee *app.CreateEmployee, updateEmployee *app.UpdateEmployee, deleteEmployee *app.DeleteEmployee) *AuthHandler {
	return &AuthHandler{
		authenticate:   authenticate,
		createEmployee: createEmployee,
		updateEmployee: updateEmployee,
		deleteEmployee: deleteEmployee,
	}
}

// Authenticate handles the authentication and JWT token generation.
func (h *AuthHandler) Authenticate(ctx context.Context, req *proto.AuthenticateRequest) (*proto.AuthenticateResponse, error) {
	token, refreshToken, err := h.authenticate.Execute(req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %v", err)
	}

	return &proto.AuthenticateResponse{
		Token:        token,
		RefreshToken: refreshToken,
		Message:      "Authentication successful",
	}, nil
}

// CreateEmployee handles the creation of a new employee by admin.
func (h *AuthHandler) CreateEmployee(ctx context.Context, req *proto.CreateEmployeeRequest) (*proto.CreateEmployeeResponse, error) {
	message, err := h.createEmployee.Execute(req.GetUsername(), req.GetPassword(), req.GetRole(), req.GetRequester())
	if err != nil {
		logging.Logger.Warn().Err(err).Str("username", req.GetUsername()).Msg("create employee failed")
		return &proto.CreateEmployeeResponse{
			Message: message,
			Success: false,
		}, nil
	}
	return &proto.CreateEmployeeResponse{
		Message: message,
		Success: true,
	}, nil
}

// UpdateRole handles updating the role for an existing employee by admin.
func (h *AuthHandler) UpdateRole(ctx context.Context, req *proto.UpdateRoleRequest) (*proto.UpdateRoleResponse, error) {
	message, err := h.updateEmployee.Execute(req.GetUsername(), req.GetRole(), req.GetRequester())
	if err != nil {
		logging.Logger.Warn().Err(err).Str("username", req.GetUsername()).Msg("update role failed")
		return &proto.UpdateRoleResponse{
			Message: "Failed to update role",
			Success: false,
		}, nil
	}
	return &proto.UpdateRoleResponse{
		Message: message,
		Success: true,
	}, nil
}

// DeleteEmployee handles the deletion (invalidation) of an employee by admin.
func (h *AuthHandler) DeleteEmployee(ctx context.Context, req *proto.DeleteEmployeeRequest) (*proto.DeleteEmployeeResponse, error) {
	message, err := h.deleteEmployee.Execute(req.GetUsername(), req.GetRequester())
	if err != nil {
		logging.Logger.Warn().Err(err).Str("username", req.GetUsername()).Msg("delete employee failed")
		return &proto.DeleteEmployeeResponse{
			Message: message,
			Success: false,
		}, nil
	}
	return &proto.DeleteEmployeeResponse{
		Message: message,
		Success: true,
	}, nil
}
