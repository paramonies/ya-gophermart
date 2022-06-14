package pgx

import "errors"

var (
	ErrConstraintViolationUser  = errors.New("login has already exist")
	ErrUserNotFound             = errors.New("user not found")
	ErrConstraintViolationOrder = errors.New("order has already registered")
	ErrOrdersNotFound           = errors.New("orders not found")
)
