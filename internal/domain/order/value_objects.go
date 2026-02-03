package order

import "errors"

// Side represents the direction of the order (buy or sell)
type Side string

const (
	SideBuy  Side = "BUY"
	SideSell Side = "SELL"
)

// IsValid checks if the side is valid
func (s Side) IsValid() bool {
	return s == SideBuy || s == SideSell
}

// String returns the string representation of Side
func (s Side) String() string {
	return string(s)
}

// OrderType represents the type of order
type OrderType string

const (
	TypeMarket    OrderType = "MARKET"     // Execute immediately at current market price
	TypeLimit     OrderType = "LIMIT"      // Execute at specified price or better
	TypeStopLoss  OrderType = "STOP_LOSS"  // Trigger market order when price reaches stop price
	TypeStopLimit OrderType = "STOP_LIMIT" // Trigger limit order when price reaches stop price
)

// IsValid checks if the order type is valid
func (t OrderType) IsValid() bool {
	switch t {
	case TypeMarket, TypeLimit, TypeStopLoss, TypeStopLimit:
		return true
	default:
		return false
	}
}

// String returns the string representation of OrderType
func (t OrderType) String() string {
	return string(t)
}

// Status represents the current status of an order
type Status string

const (
	StatusPending         Status = "PENDING"          // Order created, waiting for execution
	StatusPartiallyFilled Status = "PARTIALLY_FILLED" // Order partially executed
	StatusFilled          Status = "FILLED"           // Order fully executed
	StatusCancelled       Status = "CANCELLED"        // Order cancelled by user
	StatusRejected        Status = "REJECTED"         // Order rejected by system/broker
	StatusExpired         Status = "EXPIRED"          // Order expired (for time-limited orders)
)

// IsValid checks if the status is valid
func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusPartiallyFilled, StatusFilled, StatusCancelled, StatusRejected, StatusExpired:
		return true
	default:
		return false
	}
}

// IsFinal checks if the status is a final state (cannot be changed)
func (s Status) IsFinal() bool {
	return s == StatusFilled || s == StatusCancelled || s == StatusRejected || s == StatusExpired
}

// String returns the string representation of Status
func (s Status) String() string {
	return string(s)
}

// Domain errors
var (
	ErrOrderNotFound    = errors.New("order not found")
	ErrInvalidStatus    = errors.New("invalid order status")
	ErrInvalidOrderType = errors.New("invalid order type")
	ErrInvalidSide      = errors.New("invalid order side")
)
