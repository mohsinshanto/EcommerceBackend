package errors

import "errors"

var (
	ErrCategoryNotFound = errors.New("category not found")

	ErrCircularCategoryReference = errors.New("circular category reference detected")

	ErrExistingCircularReference = errors.New("existing circular category reference detected")
)
