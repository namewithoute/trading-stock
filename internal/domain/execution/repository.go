package execution

import "context"

// TradeRepository defines the interface for trade data access
type TradeRepository interface {
	// Create creates a new trade
	Create(ctx context.Context, trade *Trade) error

	// GetByID retrieves a trade by ID
	GetByID(ctx context.Context, id string) (*Trade, error)

	// GetByOrderID retrieves trades by order ID
	GetByOrderID(ctx context.Context, orderID string) ([]*Trade, error)

	// ListBySymbol retrieves trades by symbol
	ListBySymbol(ctx context.Context, symbol string, limit, offset int) ([]*Trade, error)

	// ListByUser retrieves trades by user ID
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*Trade, error)

	// ListByStatus retrieves trades by status
	ListByStatus(ctx context.Context, status TradeStatus, limit, offset int) ([]*Trade, error)

	// Update updates a trade
	Update(ctx context.Context, trade *Trade) error

	// CountByUser counts trades for a user
	CountByUser(ctx context.Context, userID string) (int64, error)

	// GetTotalVolumeBySymbol gets total trading volume for a symbol
	GetTotalVolumeBySymbol(ctx context.Context, symbol string) (int64, error)
}

// SettlementRepository defines the interface for settlement data access
type SettlementRepository interface {
	// Create creates a new settlement
	Create(ctx context.Context, settlement *Settlement) error

	// GetByID retrieves a settlement by ID
	GetByID(ctx context.Context, id string) (*Settlement, error)

	// GetByTradeID retrieves a settlement by trade ID
	GetByTradeID(ctx context.Context, tradeID string) (*Settlement, error)

	// ListByStatus retrieves settlements by status
	ListByStatus(ctx context.Context, status SettlementStatus, limit, offset int) ([]*Settlement, error)

	// ListByAccount retrieves settlements by account ID
	ListByAccount(ctx context.Context, accountID string, limit, offset int) ([]*Settlement, error)

	// Update updates a settlement
	Update(ctx context.Context, settlement *Settlement) error

	// ListPending retrieves all pending settlements
	ListPending(ctx context.Context, limit int) ([]*Settlement, error)

	// CountByStatus counts settlements by status
	CountByStatus(ctx context.Context, status SettlementStatus) (int64, error)
}

// ClearingRepository defines the interface for clearing instruction data access
type ClearingRepository interface {
	// Create creates a new clearing instruction
	Create(ctx context.Context, instruction *ClearingInstruction) error

	// GetByID retrieves a clearing instruction by ID
	GetByID(ctx context.Context, id string) (*ClearingInstruction, error)

	// ListByTradeID retrieves clearing instructions by trade ID
	ListByTradeID(ctx context.Context, tradeID string) ([]*ClearingInstruction, error)

	// ListByStatus retrieves clearing instructions by status
	ListByStatus(ctx context.Context, status InstructionStatus, limit, offset int) ([]*ClearingInstruction, error)

	// ListPending retrieves all pending instructions
	ListPending(ctx context.Context, limit int) ([]*ClearingInstruction, error)

	// Update updates a clearing instruction
	Update(ctx context.Context, instruction *ClearingInstruction) error

	// BatchCreate creates multiple clearing instructions
	BatchCreate(ctx context.Context, instructions []*ClearingInstruction) error

	// CountByStatus counts instructions by status
	CountByStatus(ctx context.Context, status InstructionStatus) (int64, error)
}
