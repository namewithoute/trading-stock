package execution

import (
	"time"

	domain "trading-stock/internal/domain/execution"

	"gorm.io/gorm"
)

// TradeModel is the GORM persistence model for executed trades.
type TradeModel struct {
	ID          string  `gorm:"primaryKey;type:uuid"`
	BuyOrderID  string  `gorm:"type:uuid;index;not null"`
	SellOrderID string  `gorm:"type:uuid;index;not null"`
	Symbol      string  `gorm:"type:varchar(10);index;not null"`
	Price       float64 `gorm:"type:decimal(20,4);not null"`
	Quantity    int     `gorm:"not null"`
	BuyerID     string  `gorm:"type:uuid;index;not null"`
	SellerID    string  `gorm:"type:uuid;index;not null"`
	Status      string  `gorm:"type:varchar(20);not null"`
	SettledAt   *time.Time
	CreatedAt   time.Time `gorm:"not null"`
}

func (TradeModel) TableName() string { return "trades" }

func toTradeModel(t *domain.Trade) *TradeModel {
	if t == nil {
		return nil
	}
	return &TradeModel{
		ID:          t.ID,
		BuyOrderID:  t.BuyOrderID,
		SellOrderID: t.SellOrderID,
		Symbol:      t.Symbol,
		Price:       t.Price,
		Quantity:    t.Quantity,
		BuyerID:     t.BuyerID,
		SellerID:    t.SellerID,
		Status:      string(t.Status),
		SettledAt:   t.SettledAt,
		CreatedAt:   t.CreatedAt,
	}
}

func (m *TradeModel) toDomain() *domain.Trade {
	if m == nil {
		return nil
	}
	return &domain.Trade{
		ID:          m.ID,
		BuyOrderID:  m.BuyOrderID,
		SellOrderID: m.SellOrderID,
		Symbol:      m.Symbol,
		Price:       m.Price,
		Quantity:    m.Quantity,
		BuyerID:     m.BuyerID,
		SellerID:    m.SellerID,
		Status:      domain.TradeStatus(m.Status),
		SettledAt:   m.SettledAt,
		CreatedAt:   m.CreatedAt,
	}
}

// SettlementModel is the GORM persistence model for settlements.
type SettlementModel struct {
	ID              string  `gorm:"primaryKey;type:uuid"`
	TradeID         string  `gorm:"type:uuid;uniqueIndex;not null"`
	BuyerAccountID  string  `gorm:"type:uuid;index;not null"`
	SellerAccountID string  `gorm:"type:uuid;index;not null"`
	Symbol          string  `gorm:"type:varchar(10);not null"`
	Quantity        int     `gorm:"not null"`
	Amount          float64 `gorm:"type:decimal(20,2);not null"`
	Status          string  `gorm:"type:varchar(20);not null"`
	SettledAt       *time.Time
	FailureReason   string    `gorm:"type:text"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
}

func (SettlementModel) TableName() string { return "settlements" }

func toSettlementModel(s *domain.Settlement) *SettlementModel {
	if s == nil {
		return nil
	}
	return &SettlementModel{
		ID:              s.ID,
		TradeID:         s.TradeID,
		BuyerAccountID:  s.BuyerAccountID,
		SellerAccountID: s.SellerAccountID,
		Symbol:          s.Symbol,
		Quantity:        s.Quantity,
		Amount:          s.Amount,
		Status:          string(s.Status),
		SettledAt:       s.SettledAt,
		FailureReason:   s.FailureReason,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}

func (m *SettlementModel) toDomain() *domain.Settlement {
	if m == nil {
		return nil
	}
	return &domain.Settlement{
		ID:              m.ID,
		TradeID:         m.TradeID,
		BuyerAccountID:  m.BuyerAccountID,
		SellerAccountID: m.SellerAccountID,
		Symbol:          m.Symbol,
		Quantity:        m.Quantity,
		Amount:          m.Amount,
		Status:          domain.SettlementStatus(m.Status),
		SettledAt:       m.SettledAt,
		FailureReason:   m.FailureReason,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

// ClearingInstructionModel is the GORM persistence model for clearing instructions.
type ClearingInstructionModel struct {
	ID              string  `gorm:"primaryKey;type:uuid"`
	TradeID         string  `gorm:"type:uuid;index;not null"`
	InstructionType string  `gorm:"type:varchar(20);not null"`
	FromAccountID   string  `gorm:"type:uuid;not null"`
	ToAccountID     string  `gorm:"type:uuid;not null"`
	AssetType       string  `gorm:"type:varchar(20);not null"`
	AssetSymbol     string  `gorm:"type:varchar(10)"`
	Amount          float64 `gorm:"type:decimal(20,2);not null"`
	Quantity        int
	Status          string `gorm:"type:varchar(20);not null"`
	ExecutedAt      *time.Time
	FailureReason   string    `gorm:"type:text"`
	CreatedAt       time.Time `gorm:"not null"`
}

func (ClearingInstructionModel) TableName() string { return "clearing_instructions" }

func toClearingInstructionModel(ci *domain.ClearingInstruction) *ClearingInstructionModel {
	if ci == nil {
		return nil
	}
	return &ClearingInstructionModel{
		ID:              ci.ID,
		TradeID:         ci.TradeID,
		InstructionType: string(ci.InstructionType),
		FromAccountID:   ci.FromAccountID,
		ToAccountID:     ci.ToAccountID,
		AssetType:       string(ci.AssetType),
		AssetSymbol:     ci.AssetSymbol,
		Amount:          ci.Amount,
		Quantity:        ci.Quantity,
		Status:          string(ci.Status),
		ExecutedAt:      ci.ExecutedAt,
		FailureReason:   ci.FailureReason,
		CreatedAt:       ci.CreatedAt,
	}
}

func (m *ClearingInstructionModel) toDomain() *domain.ClearingInstruction {
	if m == nil {
		return nil
	}
	return &domain.ClearingInstruction{
		ID:              m.ID,
		TradeID:         m.TradeID,
		InstructionType: domain.ClearingType(m.InstructionType),
		FromAccountID:   m.FromAccountID,
		ToAccountID:     m.ToAccountID,
		AssetType:       domain.AssetType(m.AssetType),
		AssetSymbol:     m.AssetSymbol,
		Amount:          m.Amount,
		Quantity:        m.Quantity,
		Status:          domain.InstructionStatus(m.Status),
		ExecutedAt:      m.ExecutedAt,
		FailureReason:   m.FailureReason,
		CreatedAt:       m.CreatedAt,
	}
}

// ─── Transaction helpers (used by cross-package consumers) ───────────────────

// SaveTradeWithTx persists a domain.Trade inside an already-open GORM transaction.
// Called by the Matching Service to atomically persist a trade alongside its outbox row.
func SaveTradeWithTx(tx *gorm.DB, t *domain.Trade) error {
	return tx.Create(toTradeModel(t)).Error
}

// UpdateTradeStatusWithTx updates trade status inside an already-open GORM transaction.
func UpdateTradeStatusWithTx(tx *gorm.DB, tradeID string, status domain.TradeStatus) error {
	return tx.Model(&TradeModel{}).Where("id = ?", tradeID).Update("status", string(status)).Error
}
