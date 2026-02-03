package portfolio

import "errors"

// Domain errors
var (
	ErrInsufficientQuantity = errors.New("insufficient quantity in position")
	ErrPositionNotFound     = errors.New("position not found")
	ErrInvalidQuantity      = errors.New("invalid quantity")
	ErrInvalidPrice         = errors.New("invalid price")
)
