package execution

import (
	"context"

	domain "trading-stock/internal/domain/execution"

	"gorm.io/gorm"
)

// tradeRepository implements domain.TradeRepository
type tradeRepository struct {
	db *gorm.DB
}

func NewTradeRepository(db *gorm.DB) domain.TradeRepository {
	return &tradeRepository{db: db}
}

func (r *tradeRepository) Create(ctx context.Context, t *domain.Trade) error {
	return r.db.WithContext(ctx).Create(toTradeModel(t)).Error
}

func (r *tradeRepository) GetByID(ctx context.Context, id string) (*domain.Trade, error) {
	var t TradeModel
	err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return t.toDomain(), nil
}

func (r *tradeRepository) GetByOrderID(ctx context.Context, orderID string) ([]*domain.Trade, error) {
	var models []*TradeModel
	err := r.db.WithContext(ctx).Where("buy_order_id = ? OR sell_order_id = ?", orderID, orderID).Find(&models).Error
	if err != nil {
		return nil, err
	}
	trades := make([]*domain.Trade, 0, len(models))
	for _, m := range models {
		trades = append(trades, m.toDomain())
	}
	return trades, nil
}

func (r *tradeRepository) ListBySymbol(ctx context.Context, symbol string, limit, offset int) ([]*domain.Trade, error) {
	var models []*TradeModel
	err := r.db.WithContext(ctx).Where("symbol = ?", symbol).Limit(limit).Offset(offset).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	trades := make([]*domain.Trade, 0, len(models))
	for _, m := range models {
		trades = append(trades, m.toDomain())
	}
	return trades, nil
}

func (r *tradeRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Trade, error) {
	var models []*TradeModel
	err := r.db.WithContext(ctx).Where("buyer_id = ? OR seller_id = ?", userID, userID).Limit(limit).Offset(offset).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	trades := make([]*domain.Trade, 0, len(models))
	for _, m := range models {
		trades = append(trades, m.toDomain())
	}
	return trades, nil
}

func (r *tradeRepository) ListByStatus(ctx context.Context, status domain.TradeStatus, limit, offset int) ([]*domain.Trade, error) {
	var models []*TradeModel
	err := r.db.WithContext(ctx).Where("status = ?", status).Limit(limit).Offset(offset).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	trades := make([]*domain.Trade, 0, len(models))
	for _, m := range models {
		trades = append(trades, m.toDomain())
	}
	return trades, nil
}

func (r *tradeRepository) Update(ctx context.Context, t *domain.Trade) error {
	return r.db.WithContext(ctx).Save(toTradeModel(t)).Error
}

func (r *tradeRepository) CountByUser(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&TradeModel{}).Where("buyer_id = ? OR seller_id = ?", userID, userID).Count(&count).Error
	return count, err
}

func (r *tradeRepository) GetTotalVolumeBySymbol(ctx context.Context, symbol string) (int64, error) {
	var totalVolume int64
	err := r.db.WithContext(ctx).Model(&TradeModel{}).Where("symbol = ?", symbol).Select("SUM(quantity)").Scan(&totalVolume).Error
	return totalVolume, err
}

// settlementRepository implements domain.SettlementRepository
type settlementRepository struct {
	db *gorm.DB
}

func NewSettlementRepository(db *gorm.DB) domain.SettlementRepository {
	return &settlementRepository{db: db}
}

func (r *settlementRepository) Create(ctx context.Context, s *domain.Settlement) error {
	return r.db.WithContext(ctx).Create(toSettlementModel(s)).Error
}

func (r *settlementRepository) GetByID(ctx context.Context, id string) (*domain.Settlement, error) {
	var s SettlementModel
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return s.toDomain(), nil
}

func (r *settlementRepository) GetByTradeID(ctx context.Context, tradeID string) (*domain.Settlement, error) {
	var s SettlementModel
	err := r.db.WithContext(ctx).Where("trade_id = ?", tradeID).First(&s).Error
	if err != nil {
		return nil, err
	}
	return s.toDomain(), nil
}

func (r *settlementRepository) ListByStatus(ctx context.Context, status domain.SettlementStatus, limit, offset int) ([]*domain.Settlement, error) {
	var models []*SettlementModel
	err := r.db.WithContext(ctx).Where("status = ?", status).Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, err
	}
	settlements := make([]*domain.Settlement, 0, len(models))
	for _, m := range models {
		settlements = append(settlements, m.toDomain())
	}
	return settlements, nil
}

func (r *settlementRepository) ListByAccount(ctx context.Context, accountID string, limit, offset int) ([]*domain.Settlement, error) {
	var models []*SettlementModel
	err := r.db.WithContext(ctx).Where("buyer_account_id = ? OR seller_account_id = ?", accountID, accountID).Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, err
	}
	settlements := make([]*domain.Settlement, 0, len(models))
	for _, m := range models {
		settlements = append(settlements, m.toDomain())
	}
	return settlements, nil
}

func (r *settlementRepository) Update(ctx context.Context, s *domain.Settlement) error {
	return r.db.WithContext(ctx).Save(toSettlementModel(s)).Error
}

func (r *settlementRepository) ListPending(ctx context.Context, limit int) ([]*domain.Settlement, error) {
	var models []*SettlementModel
	err := r.db.WithContext(ctx).Where("status = ?", domain.SettlementStatusPending).Limit(limit).Find(&models).Error
	if err != nil {
		return nil, err
	}
	settlements := make([]*domain.Settlement, 0, len(models))
	for _, m := range models {
		settlements = append(settlements, m.toDomain())
	}
	return settlements, nil
}

func (r *settlementRepository) CountByStatus(ctx context.Context, status domain.SettlementStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&SettlementModel{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// clearingRepository implements domain.ClearingRepository
type clearingRepository struct {
	db *gorm.DB
}

func NewClearingRepository(db *gorm.DB) domain.ClearingRepository {
	return &clearingRepository{db: db}
}

func (r *clearingRepository) Create(ctx context.Context, ci *domain.ClearingInstruction) error {
	return r.db.WithContext(ctx).Create(toClearingInstructionModel(ci)).Error
}

func (r *clearingRepository) GetByID(ctx context.Context, id string) (*domain.ClearingInstruction, error) {
	var ci ClearingInstructionModel
	err := r.db.WithContext(ctx).First(&ci, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return ci.toDomain(), nil
}

func (r *clearingRepository) ListByTradeID(ctx context.Context, tradeID string) ([]*domain.ClearingInstruction, error) {
	var models []*ClearingInstructionModel
	err := r.db.WithContext(ctx).Where("trade_id = ?", tradeID).Find(&models).Error
	if err != nil {
		return nil, err
	}
	instructions := make([]*domain.ClearingInstruction, 0, len(models))
	for _, m := range models {
		instructions = append(instructions, m.toDomain())
	}
	return instructions, nil
}

func (r *clearingRepository) ListByStatus(ctx context.Context, status domain.InstructionStatus, limit, offset int) ([]*domain.ClearingInstruction, error) {
	var models []*ClearingInstructionModel
	err := r.db.WithContext(ctx).Where("status = ?", status).Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, err
	}
	instructions := make([]*domain.ClearingInstruction, 0, len(models))
	for _, m := range models {
		instructions = append(instructions, m.toDomain())
	}
	return instructions, nil
}

func (r *clearingRepository) ListPending(ctx context.Context, limit int) ([]*domain.ClearingInstruction, error) {
	var models []*ClearingInstructionModel
	err := r.db.WithContext(ctx).Where("status = ?", domain.InstructionStatusPending).Limit(limit).Find(&models).Error
	if err != nil {
		return nil, err
	}
	instructions := make([]*domain.ClearingInstruction, 0, len(models))
	for _, m := range models {
		instructions = append(instructions, m.toDomain())
	}
	return instructions, nil
}

func (r *clearingRepository) Update(ctx context.Context, ci *domain.ClearingInstruction) error {
	return r.db.WithContext(ctx).Save(toClearingInstructionModel(ci)).Error
}

func (r *clearingRepository) BatchCreate(ctx context.Context, instructions []*domain.ClearingInstruction) error {
	models := make([]*ClearingInstructionModel, 0, len(instructions))
	for _, ci := range instructions {
		models = append(models, toClearingInstructionModel(ci))
	}
	return r.db.WithContext(ctx).Create(&models).Error
}

func (r *clearingRepository) CountByStatus(ctx context.Context, status domain.InstructionStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&ClearingInstructionModel{}).Where("status = ?", status).Count(&count).Error
	return count, err
}
