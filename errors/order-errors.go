package errors

import "errors"

var (
	ErrFailedToStartTransaction = errors.New("failed to start transaction")
	ErrFailedToCreateOrder      = errors.New("failed to create order")
	ErrFailedToCreateOrderItems = errors.New("failed to create order item")
	ErrTransactionFailed        = errors.New("transaction failed")
)
