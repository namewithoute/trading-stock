package portfolio

import "time"

// Position represents a current holding in a portfolio
// This tracks the user's current positions in various securities
type Position struct {
	ID        string `json:"id" gorm:"primaryKey;type:uuid"`
	UserID    string `json:"user_id" gorm:"type:uuid;index;not null"`
	AccountID string `json:"account_id" gorm:"type:uuid;index;not null"`
	Symbol    string `json:"symbol" gorm:"type:varchar(10);index;not null"`

	// Position details
	Quantity     int     `json:"quantity" gorm:"not null"`
	AvgPrice     float64 `json:"avg_price" gorm:"type:decimal(20,4);not null"`
	CurrentPrice float64 `json:"current_price" gorm:"type:decimal(20,4)"`

	// P&L calculations
	UnrealizedPnL        float64 `json:"unrealized_pnl" gorm:"type:decimal(20,2)"`
	UnrealizedPnLPercent float64 `json:"unrealized_pnl_percent" gorm:"type:decimal(10,4)"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
}

// TableName specifies the table name for GORM
func (Position) TableName() string {
	return "positions"
}

// TotalCost returns the total cost basis of the position
func (p *Position) TotalCost() float64 {
	return p.AvgPrice * float64(p.Quantity)
}

// CurrentValue returns the current market value of the position
func (p *Position) CurrentValue() float64 {
	return p.CurrentPrice * float64(p.Quantity)
}

// CalculateUnrealizedPnL calculates the unrealized profit/loss
func (p *Position) CalculateUnrealizedPnL() {
	p.UnrealizedPnL = p.CurrentValue() - p.TotalCost()
	if p.TotalCost() > 0 {
		p.UnrealizedPnLPercent = (p.UnrealizedPnL / p.TotalCost()) * 100
	}
}

// UpdateCurrentPrice updates the current price and recalculates P&L
func (p *Position) UpdateCurrentPrice(price float64) {
	p.CurrentPrice = price
	p.CalculateUnrealizedPnL()
	p.UpdatedAt = time.Now()
}

// AddQuantity adds to the position (buying more)
func (p *Position) AddQuantity(quantity int, price float64) {
	totalCost := p.TotalCost() + (float64(quantity) * price)
	p.Quantity += quantity
	p.AvgPrice = totalCost / float64(p.Quantity)
	p.CalculateUnrealizedPnL()
	p.UpdatedAt = time.Now()
}

// ReduceQuantity reduces the position (selling)
func (p *Position) ReduceQuantity(quantity int) error {
	if quantity > p.Quantity {
		return ErrInsufficientQuantity
	}
	p.Quantity -= quantity
	p.CalculateUnrealizedPnL()
	p.UpdatedAt = time.Now()
	return nil
}

// IsClosed checks if the position is closed (quantity = 0)
func (p *Position) IsClosed() bool {
	return p.Quantity == 0
}
