package execution

import "errors"

// TradeStatus represents the status of a trade
type TradeStatus string

const (
	TradeStatusPending   TradeStatus = "PENDING"
	TradeStatusSettled   TradeStatus = "SETTLED"
	TradeStatusFailed    TradeStatus = "FAILED"
	TradeStatusCancelled TradeStatus = "CANCELLED"
)

// IsValid checks if the trade status is valid
func (ts TradeStatus) IsValid() bool {
	switch ts {
	case TradeStatusPending, TradeStatusSettled, TradeStatusFailed, TradeStatusCancelled:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (ts TradeStatus) String() string {
	return string(ts)
}

// SettlementStatus represents the status of a settlement
type SettlementStatus string

const (
	SettlementStatusPending   SettlementStatus = "PENDING"
	SettlementStatusCompleted SettlementStatus = "COMPLETED"
	SettlementStatusFailed    SettlementStatus = "FAILED"
)

// IsValid checks if the settlement status is valid
func (ss SettlementStatus) IsValid() bool {
	switch ss {
	case SettlementStatusPending, SettlementStatusCompleted, SettlementStatusFailed:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (ss SettlementStatus) String() string {
	return string(ss)
}

// ClearingType represents the type of clearing instruction
type ClearingType string

const (
	ClearingTypeCash  ClearingType = "CASH"
	ClearingTypeStock ClearingType = "STOCK"
)

// IsValid checks if the clearing type is valid
func (ct ClearingType) IsValid() bool {
	switch ct {
	case ClearingTypeCash, ClearingTypeStock:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (ct ClearingType) String() string {
	return string(ct)
}

// AssetType represents the type of asset
type AssetType string

const (
	AssetTypeCash  AssetType = "CASH"
	AssetTypeStock AssetType = "STOCK"
)

// IsValid checks if the asset type is valid
func (at AssetType) IsValid() bool {
	switch at {
	case AssetTypeCash, AssetTypeStock:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (at AssetType) String() string {
	return string(at)
}

// InstructionStatus represents the status of a clearing instruction
type InstructionStatus string

const (
	InstructionStatusPending  InstructionStatus = "PENDING"
	InstructionStatusExecuted InstructionStatus = "EXECUTED"
	InstructionStatusFailed   InstructionStatus = "FAILED"
)

// IsValid checks if the instruction status is valid
func (is InstructionStatus) IsValid() bool {
	switch is {
	case InstructionStatusPending, InstructionStatusExecuted, InstructionStatusFailed:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (is InstructionStatus) String() string {
	return string(is)
}

// Domain errors
var (
	ErrTradeAlreadySettled        = errors.New("trade already settled")
	ErrSettlementAlreadyCompleted = errors.New("settlement already completed")
	ErrSettlementFailed           = errors.New("settlement failed")
	ErrInstructionAlreadyExecuted = errors.New("instruction already executed")
	ErrInsufficientFunds          = errors.New("insufficient funds for settlement")
	ErrInsufficientShares         = errors.New("insufficient shares for settlement")
)
