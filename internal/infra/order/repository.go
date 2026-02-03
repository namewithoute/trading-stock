package order

import (
	"context"
	"errors"

	"trading-stock/internal/domain/order"

	"gorm.io/gorm"
)

type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new order repository implementation
func NewOrderRepository(db *gorm.DB) order.Repository {
	return &orderRepository{db: db}
}

// Create creates a new order in the database
func (r *orderRepository) Create(ctx context.Context, o *order.Order) error {
	return r.db.WithContext(ctx).Create(o).Error
}

// GetByID retrieves an order by its ID
func (r *orderRepository) GetByID(ctx context.Context, id string) (*order.Order, error) {
	var o order.Order
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&o).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}
	return &o, nil
}

// Update updates an existing order
func (r *orderRepository) Update(ctx context.Context, o *order.Order) error {
	return r.db.WithContext(ctx).Save(o).Error
}

// Cancel cancels an order by setting its status to CANCELLED
func (r *orderRepository) Cancel(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&order.Order{}).
		Where("id = ?", id).
		Update("status", order.StatusCancelled).
		Error
}

// ListByUserID retrieves all orders for a specific user
func (r *orderRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*order.Order, error) {
	var orders []*order.Order
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&orders).Error
	return orders, err
}

// ListByStatus retrieves all orders with a specific status
func (r *orderRepository) ListByStatus(ctx context.Context, status order.Status, limit, offset int) ([]*order.Order, error) {
	var orders []*order.Order
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&orders).Error
	return orders, err
}

// ListBySymbol retrieves all orders for a specific symbol
func (r *orderRepository) ListBySymbol(ctx context.Context, symbol string, limit, offset int) ([]*order.Order, error) {
	var orders []*order.Order
	err := r.db.WithContext(ctx).
		Where("symbol = ?", symbol).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&orders).Error
	return orders, err
}

// ListPendingOrdersByUserAndSymbol retrieves pending orders for a user and symbol
func (r *orderRepository) ListPendingOrdersByUserAndSymbol(ctx context.Context, userID, symbol string) ([]*order.Order, error) {
	var orders []*order.Order
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND symbol = ? AND status IN ?", userID, symbol, []order.Status{order.StatusPending, order.StatusPartiallyFilled}).
		Order("created_at DESC").
		Find(&orders).Error
	return orders, err
}

// CountByUserID returns the total number of orders for a user
func (r *orderRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&order.Order{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// Delete permanently deletes an order
func (r *orderRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&order.Order{}).Error
}
