package execution

import (
	"time"

	"github.com/cockroachdb/apd/v3"
)

var decCtx = apd.BaseContext.WithPrecision(19)

// Trade represents an executed trade between a buyer and seller
type Trade struct {
	ID          string
	BuyOrderID  string
	SellOrderID string
	Symbol      string
	Price       apd.Decimal
	Quantity    int
	BuyerID     string
	SellerID    string
	Status      TradeStatus
	SettledAt   *time.Time
	CreatedAt   time.Time
}

// TotalValue returns the total value of the trade
func (t *Trade) TotalValue() apd.Decimal {
	var result apd.Decimal
	_, _ = decCtx.Mul(&result, &t.Price, apd.New(int64(t.Quantity), 0))
	return result
}

// IsSettled checks if the trade has been settled
func (t *Trade) IsSettled() bool {
	return t.Status == TradeStatusSettled
}

// CanSettle checks if the trade can be settled
func (t *Trade) CanSettle() bool {
	return t.Status == TradeStatusPending
}

// Settle marks the trade as settled
func (t *Trade) Settle() error {
	if !t.CanSettle() {
		return ErrTradeAlreadySettled
	}

	now := time.Now()
	t.Status = TradeStatusSettled
	t.SettledAt = &now
	return nil
}

// Fail marks the trade as failed
func (t *Trade) Fail() error {
	if t.IsSettled() {
		return ErrTradeAlreadySettled
	}

	t.Status = TradeStatusFailed
	return nil
}

// Settlement represents the settlement of a trade
type Settlement struct {
	ID              string
	TradeID         string
	BuyerAccountID  string
	SellerAccountID string
	Symbol          string
	Quantity        int
	Amount          apd.Decimal
	Status          SettlementStatus
	SettledAt       *time.Time
	FailureReason   string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// IsCompleted checks if settlement is completed
func (s *Settlement) IsCompleted() bool {
	return s.Status == SettlementStatusCompleted
}

// IsFailed checks if settlement failed
func (s *Settlement) IsFailed() bool {
	return s.Status == SettlementStatusFailed
}

// Complete marks the settlement as completed
func (s *Settlement) Complete() error {
	if s.IsCompleted() {
		return ErrSettlementAlreadyCompleted
	}

	if s.IsFailed() {
		return ErrSettlementFailed
	}

	now := time.Now()
	s.Status = SettlementStatusCompleted
	s.SettledAt = &now
	return nil
}

// Fail marks the settlement as failed with a reason
func (s *Settlement) Fail(reason string) error {
	if s.IsCompleted() {
		return ErrSettlementAlreadyCompleted
	}

	s.Status = SettlementStatusFailed
	s.FailureReason = reason
	return nil
}

// ClearingInstruction represents instructions for clearing a trade
type ClearingInstruction struct {
	ID              string
	TradeID         string
	InstructionType ClearingType
	FromAccountID   string
	ToAccountID     string
	AssetType       AssetType
	AssetSymbol     string
	Amount          apd.Decimal
	Quantity        int
	Status          InstructionStatus
	ExecutedAt      *time.Time
	FailureReason   string
	CreatedAt       time.Time
}

// IsExecuted checks if the instruction has been executed
func (ci *ClearingInstruction) IsExecuted() bool {
	return ci.Status == InstructionStatusExecuted
}

// CanExecute checks if the instruction can be executed
func (ci *ClearingInstruction) CanExecute() bool {
	return ci.Status == InstructionStatusPending
}

// Execute marks the instruction as executed
func (ci *ClearingInstruction) Execute() error {
	if !ci.CanExecute() {
		return ErrInstructionAlreadyExecuted
	}

	now := time.Now()
	ci.Status = InstructionStatusExecuted
	ci.ExecutedAt = &now
	return nil
}

// Fail marks the instruction as failed
func (ci *ClearingInstruction) Fail(reason string) error {
	if ci.IsExecuted() {
		return ErrInstructionAlreadyExecuted
	}

	ci.Status = InstructionStatusFailed
	ci.FailureReason = reason
	return nil
}
