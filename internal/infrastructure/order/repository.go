package order

import (
	"context"
	"errors"

	domain "trading-stock/internal/domain/order"

	"gorm.io/gorm"
)

type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new order repository implementation
func NewOrderRepository(db *gorm.DB) domain.Repository {
	return &orderRepository{db: db}
}

// Create creates a new order in the database
func (r *orderRepository) Create(ctx context.Context, o *domain.Order) error {
	return r.db.WithContext(ctx).Create(toOrderModel(o)).Error
}

// GetByID retrieves an order by its ID
func (r *orderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	var o OrderModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&o).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}
	return o.toDomain(), nil
}

// Update updates an existing order
func (r *orderRepository) Update(ctx context.Context, o *domain.Order) error {
	return r.db.WithContext(ctx).Save(toOrderModel(o)).Error
}

// Cancel cancels an order by setting its status to CANCELLED
func (r *orderRepository) Cancel(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&OrderModel{}).
		Where("id = ?", id).
		Update("status", domain.StatusCancelled).
		Error
}

// ListByUserID retrieves all orders for a specific user
func (r *orderRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Order, error) {
	var models []*OrderModel
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	orders := make([]*domain.Order, 0, len(models))
	for _, m := range models {
		orders = append(orders, m.toDomain())
	}
	return orders, nil
}

// ListByStatus retrieves all orders with a specific status
func (r *orderRepository) ListByStatus(ctx context.Context, status domain.Status, limit, offset int) ([]*domain.Order, error) {
	var models []*OrderModel
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	orders := make([]*domain.Order, 0, len(models))
	for _, m := range models {
		orders = append(orders, m.toDomain())
	}
	return orders, nil
}

// ListBySymbol retrieves all orders for a specific symbol
func (r *orderRepository) ListBySymbol(ctx context.Context, symbol string, limit, offset int) ([]*domain.Order, error) {
	var models []*OrderModel
	err := r.db.WithContext(ctx).
		Where("symbol = ?", symbol).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	orders := make([]*domain.Order, 0, len(models))
	for _, m := range models {
		orders = append(orders, m.toDomain())
	}
	return orders, nil
}

// ListPendingOrdersByUserAndSymbol retrieves pending orders for a user and symbol
func (r *orderRepository) ListPendingOrdersByUserAndSymbol(ctx context.Context, userID, symbol string) ([]*domain.Order, error) {
	var models []*OrderModel
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND symbol = ? AND status IN ?", userID, symbol, []domain.Status{domain.StatusPending, domain.StatusPartiallyFilled}).
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	orders := make([]*domain.Order, 0, len(models))
	for _, m := range models {
		orders = append(orders, m.toDomain())
	}
	return orders, nil
}

// CountByUserID returns the total number of orders for a user
func (r *orderRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&OrderModel{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// Delete permanently deletes an order
func (r *orderRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&OrderModel{}).Error
}
