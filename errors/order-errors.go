package errors

import "errors"

var (
	ErrFailedToStartTransaction = errors.New("failed to start transaction")
	ErrFailedToCreateOrder      = errors.New("failed to create order")
	ErrFailedToCreateOrderItems = errors.New("failed to create order item")
	ErrTransactionFailed        = errors.New("transaction failed")

	ErrSSLCommerzNotConfigured  = errors.New("sslcommerz is not configured")
	ErrPaymentSessionUnknown    = errors.New("payment session result is unknown")
	ErrPaymentSessionSaveFailed = errors.New("payment session details could not be saved")

	ErrOrderNotFound = errors.New("order not found")

	// Payment validation errors
	ErrMissingPaymentInformation   = errors.New("missing payment transaction information")
	ErrPaymentValidationFailed     = errors.New("payment validation failed")
	ErrPaymentTransactionMismatch  = errors.New("payment transaction ID mismatch")
	ErrPaymentValidationIDMismatch = errors.New("payment validation ID mismatch")
	ErrInvalidPaymentAmount        = errors.New("invalid payment amount")
	ErrPaymentAmountMismatch       = errors.New("payment amount mismatch")
	ErrPaymentCurrencyMismatch     = errors.New("payment currency mismatch")
	ErrInvalidPaymentMethod        = errors.New("order does not use SSLCommerz")
	ErrFailedToUpdatePayment       = errors.New("failed to update payment status")
)
