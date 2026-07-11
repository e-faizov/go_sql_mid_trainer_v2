package domain

import "errors"

var (
	ErrNotImplemented     = errors.New("not implemented")
	ErrWrongID            = errors.New("wrong id")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUserNotFound       = errors.New("user not found")
	ErrOrderNotFound      = errors.New("order not found")
	ErrConflict           = errors.New("conflict")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrExternalBadStatus  = errors.New("external service bad status")
	ErrIdempotencyMissing = errors.New("idempotency key is required")
)
