package error

import "errors"

var (
	ErrValidationFailed = errors.New("validation failed")
	ErrInvalidUsername  = errors.New("invalid username")
)
