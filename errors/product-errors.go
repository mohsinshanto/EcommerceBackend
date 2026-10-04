package errors

import "errors"

var (
	ErrProductNotFound     = errors.New("product not found")
	ErrInsufficientStock   = errors.New("Insufficient stock")
	ErrFailedToUpdateStock = errors.New("failed to update product stock")
)
