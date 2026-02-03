package execution

import (
	"time"
)

// Trade represents an executed trade between a buyer and seller
type Trade struct {
	ID          string      `json:"id" gorm:"primaryKey;type:uuid"`
	BuyOrderID  string      `json:"buy_order_id" gorm:"type:uuid;index;not null"`
	SellOrderID string      `json:"sell_order_id" gorm:"type:uuid;index;not null"`
	Symbol      string      `json:"symbol" gorm:"type:varchar(10);index;not null"`
	Price       float64     `json:"price" gorm:"type:decimal(20,4);not null"`
	Quantity    int         `json:"quantity" gorm:"not null"`
	BuyerID     string      `json:"buyer_id" gorm:"type:uuid;index;not null"`
	SellerID    string      `json:"seller_id" gorm:"type:uuid;index;not null"`
	Status      TradeStatus `json:"status" gorm:"type:varchar(20);not null"`
	SettledAt   *time.Time  `json:"settled_at,omitempty"`
	CreatedAt   time.Time   `json:"created_at" gorm:"not null"`
}

// TotalValue returns the total value of the trade
func (t *Trade) TotalValue() float64 {
	return t.Price * float64(t.Quantity)
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
	ID              string           `json:"id" gorm:"primaryKey;type:uuid"`
	TradeID         string           `json:"trade_id" gorm:"type:uuid;uniqueIndex;not null"`
	BuyerAccountID  string           `json:"buyer_account_id" gorm:"type:uuid;index;not null"`
	SellerAccountID string           `json:"seller_account_id" gorm:"type:uuid;index;not null"`
	Symbol          string           `json:"symbol" gorm:"type:varchar(10);not null"`
	Quantity        int              `json:"quantity" gorm:"not null"`
	Amount          float64          `json:"amount" gorm:"type:decimal(20,2);not null"`
	Status          SettlementStatus `json:"status" gorm:"type:varchar(20);not null"`
	SettledAt       *time.Time       `json:"settled_at,omitempty"`
	FailureReason   string           `json:"failure_reason,omitempty" gorm:"type:text"`
	CreatedAt       time.Time        `json:"created_at" gorm:"not null"`
	UpdatedAt       time.Time        `json:"updated_at" gorm:"not null"`
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
	ID              string            `json:"id" gorm:"primaryKey;type:uuid"`
	TradeID         string            `json:"trade_id" gorm:"type:uuid;index;not null"`
	InstructionType ClearingType      `json:"instruction_type" gorm:"type:varchar(20);not null"`
	FromAccountID   string            `json:"from_account_id" gorm:"type:uuid;not null"`
	ToAccountID     string            `json:"to_account_id" gorm:"type:uuid;not null"`
	AssetType       AssetType         `json:"asset_type" gorm:"type:varchar(20);not null"`
	AssetSymbol     string            `json:"asset_symbol,omitempty" gorm:"type:varchar(10)"`
	Amount          float64           `json:"amount" gorm:"type:decimal(20,2);not null"`
	Quantity        int               `json:"quantity,omitempty"`
	Status          InstructionStatus `json:"status" gorm:"type:varchar(20);not null"`
	ExecutedAt      *time.Time        `json:"executed_at,omitempty"`
	FailureReason   string            `json:"failure_reason,omitempty" gorm:"type:text"`
	CreatedAt       time.Time         `json:"created_at" gorm:"not null"`
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
