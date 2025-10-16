package error

import "errors"

var (
	ErrValidationFailed      = errors.New("validation failed")
	ErrInvalidUsername       = errors.New("invalid username")
	ErrInvalidRequest        = errors.New("invalid request")
	ErrInvalidRole           = errors.New("invalid role")
	ErrMissingRequiredData   = errors.New("missing required data")
	ErrEmployeeAlreadyExists = errors.New("employee already exists")
	ErrEmployeeNotFound      = errors.New("employee not found")
	ErrDatabase              = errors.New("database error")
)
