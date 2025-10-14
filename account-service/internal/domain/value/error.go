package value

import "errors"

var (
	ErrCustomerExists      = errors.New("customer already exists")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrInvalidCustomer     = errors.New("invalid customer data")
	ErrMissingCustomerID   = errors.New("customer ID is required")
	ErrNegativeBalance     = errors.New("balance cannot be negative")
	ErrSameAccountTransfer = errors.New("cannot transfer to same account")
	ErrMissingReferenceID  = errors.New("reference ID is required")
	ErrValidationFailed    = errors.New("validation failed")
	ErrAccountExists       = errors.New("account already exists")
	ErrDatabase            = errors.New("database error")
	ErrTimeout             = errors.New("operation timeout")
	ErrCustomerNotFound    = errors.New("customer not found")
	ErrCustomerLocked      = errors.New("customer is locked for operation")
	ErrCustomerHasActiveTx = errors.New("customer has accounts with active transactions")
)
