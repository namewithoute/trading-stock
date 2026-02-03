package execution

import (
	"context"

	"trading-stock/internal/domain/execution"

	"gorm.io/gorm"
)

// tradeRepository implements domain.TradeRepository
type tradeRepository struct {
	db *gorm.DB
}

func NewTradeRepository(db *gorm.DB) execution.TradeRepository {
	return &tradeRepository{db: db}
}

func (r *tradeRepository) Create(ctx context.Context, t *execution.Trade) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *tradeRepository) GetByID(ctx context.Context, id string) (*execution.Trade, error) {
	var t execution.Trade
	err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error
	return &t, err
}

func (r *tradeRepository) GetByOrderID(ctx context.Context, orderID string) ([]*execution.Trade, error) {
	var trades []*execution.Trade
	err := r.db.WithContext(ctx).Where("buy_order_id = ? OR sell_order_id = ?", orderID, orderID).Find(&trades).Error
	return trades, err
}

func (r *tradeRepository) ListBySymbol(ctx context.Context, symbol string, limit, offset int) ([]*execution.Trade, error) {
	var trades []*execution.Trade
	err := r.db.WithContext(ctx).Where("symbol = ?", symbol).Limit(limit).Offset(offset).Order("created_at DESC").Find(&trades).Error
	return trades, err
}

func (r *tradeRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*execution.Trade, error) {
	var trades []*execution.Trade
	err := r.db.WithContext(ctx).Where("buyer_id = ? OR seller_id = ?", userID, userID).Limit(limit).Offset(offset).Order("created_at DESC").Find(&trades).Error
	return trades, err
}

func (r *tradeRepository) ListByStatus(ctx context.Context, status execution.TradeStatus, limit, offset int) ([]*execution.Trade, error) {
	var trades []*execution.Trade
	err := r.db.WithContext(ctx).Where("status = ?", status).Limit(limit).Offset(offset).Order("created_at DESC").Find(&trades).Error
	return trades, err
}

func (r *tradeRepository) Update(ctx context.Context, t *execution.Trade) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *tradeRepository) CountByUser(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&execution.Trade{}).Where("buyer_id = ? OR seller_id = ?", userID, userID).Count(&count).Error
	return count, err
}

func (r *tradeRepository) GetTotalVolumeBySymbol(ctx context.Context, symbol string) (int64, error) {
	var totalVolume int64
	err := r.db.WithContext(ctx).Model(&execution.Trade{}).Where("symbol = ?", symbol).Select("SUM(quantity)").Scan(&totalVolume).Error
	return totalVolume, err
}

// settlementRepository implements domain.SettlementRepository
type settlementRepository struct {
	db *gorm.DB
}

func NewSettlementRepository(db *gorm.DB) execution.SettlementRepository {
	return &settlementRepository{db: db}
}

func (r *settlementRepository) Create(ctx context.Context, s *execution.Settlement) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *settlementRepository) GetByID(ctx context.Context, id string) (*execution.Settlement, error) {
	var s execution.Settlement
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	return &s, err
}

func (r *settlementRepository) GetByTradeID(ctx context.Context, tradeID string) (*execution.Settlement, error) {
	var s execution.Settlement
	err := r.db.WithContext(ctx).Where("trade_id = ?", tradeID).First(&s).Error
	return &s, err
}

func (r *settlementRepository) ListByStatus(ctx context.Context, status execution.SettlementStatus, limit, offset int) ([]*execution.Settlement, error) {
	var settlements []*execution.Settlement
	err := r.db.WithContext(ctx).Where("status = ?", status).Limit(limit).Offset(offset).Find(&settlements).Error
	return settlements, err
}

func (r *settlementRepository) ListByAccount(ctx context.Context, accountID string, limit, offset int) ([]*execution.Settlement, error) {
	var settlements []*execution.Settlement
	err := r.db.WithContext(ctx).Where("buyer_account_id = ? OR seller_account_id = ?", accountID, accountID).Limit(limit).Offset(offset).Find(&settlements).Error
	return settlements, err
}

func (r *settlementRepository) Update(ctx context.Context, s *execution.Settlement) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *settlementRepository) ListPending(ctx context.Context, limit int) ([]*execution.Settlement, error) {
	var settlements []*execution.Settlement
	err := r.db.WithContext(ctx).Where("status = ?", execution.SettlementStatusPending).Limit(limit).Find(&settlements).Error
	return settlements, err
}

func (r *settlementRepository) CountByStatus(ctx context.Context, status execution.SettlementStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&execution.Settlement{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// clearingRepository implements domain.ClearingRepository
type clearingRepository struct {
	db *gorm.DB
}

func NewClearingRepository(db *gorm.DB) execution.ClearingRepository {
	return &clearingRepository{db: db}
}

func (r *clearingRepository) Create(ctx context.Context, ci *execution.ClearingInstruction) error {
	return r.db.WithContext(ctx).Create(ci).Error
}

func (r *clearingRepository) GetByID(ctx context.Context, id string) (*execution.ClearingInstruction, error) {
	var ci execution.ClearingInstruction
	err := r.db.WithContext(ctx).First(&ci, "id = ?", id).Error
	return &ci, err
}

func (r *clearingRepository) ListByTradeID(ctx context.Context, tradeID string) ([]*execution.ClearingInstruction, error) {
	var instructions []*execution.ClearingInstruction
	err := r.db.WithContext(ctx).Where("trade_id = ?", tradeID).Find(&instructions).Error
	return instructions, err
}

func (r *clearingRepository) ListByStatus(ctx context.Context, status execution.InstructionStatus, limit, offset int) ([]*execution.ClearingInstruction, error) {
	var instructions []*execution.ClearingInstruction
	err := r.db.WithContext(ctx).Where("status = ?", status).Limit(limit).Offset(offset).Find(&instructions).Error
	return instructions, err
}

func (r *clearingRepository) ListPending(ctx context.Context, limit int) ([]*execution.ClearingInstruction, error) {
	var instructions []*execution.ClearingInstruction
	err := r.db.WithContext(ctx).Where("status = ?", execution.InstructionStatusPending).Limit(limit).Find(&instructions).Error
	return instructions, err
}

func (r *clearingRepository) Update(ctx context.Context, ci *execution.ClearingInstruction) error {
	return r.db.WithContext(ctx).Save(ci).Error
}

func (r *clearingRepository) BatchCreate(ctx context.Context, instructions []*execution.ClearingInstruction) error {
	return r.db.WithContext(ctx).Create(&instructions).Error
}

func (r *clearingRepository) CountByStatus(ctx context.Context, status execution.InstructionStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&execution.ClearingInstruction{}).Where("status = ?", status).Count(&count).Error
	return count, err
}
