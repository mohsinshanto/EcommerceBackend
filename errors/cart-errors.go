package errors

import "errors"

var (
	ErrUpdateCart         = errors.New("failed to update cart")
	ErrFailedToAddCart    = errors.New("failed to add in the cart")
	ErrFailedToLoadCart   = errors.New("failed to load cart")
	ErrCartNotFound       = errors.New("cart not found")
	ErrQuantityValidation = errors.New("quantity must be greater than zero")
	ErrRemoveCartItem     = errors.New("failed to remove cart item")
	ErrForbidden          = errors.New("forbidden")
	ErrCartEmpty          = errors.New("cart is empty")
	ErrFailedToClearCart  = errors.New("failed to create cart")
)
