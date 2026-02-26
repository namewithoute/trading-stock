package order

import (
	"context"
	"fmt"
	"time"

	domain "trading-stock/internal/domain/order"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────────────────────────────────
// GORM Models
// ─────────────────────────────────────────────────────────────────────────────

// OrderEventModel maps to `order_events` — the immutable append-only log.
type OrderEventModel struct {
	ID          string    `gorm:"primaryKey;type:uuid"`
	AggregateID string    `gorm:"type:uuid;index;not null"`
	EventType   string    `gorm:"type:varchar(64);not null"`
	Payload     []byte    `gorm:"type:jsonb;not null"`
	Version     int       `gorm:"not null"`
	OccurredAt  time.Time `gorm:"not null"`
}

func (OrderEventModel) TableName() string { return "order_events" }

// OrderReadModelDB maps to `order_read_models` — the denormalised query table.
type OrderReadModelDB struct {
	ID             string    `gorm:"primaryKey;type:uuid"`
	UserID         string    `gorm:"type:uuid;index;not null"`
	AccountID      string    `gorm:"type:uuid;index"`
	Symbol         string    `gorm:"type:varchar(10);index;not null"`
	Side           string    `gorm:"type:varchar(10);not null"`
	OrderType      string    `gorm:"column:order_type;type:varchar(20);not null"`
	Quantity       int       `gorm:"not null"`
	Price          float64   `gorm:"type:decimal(20,4);not null"`
	FilledQuantity int       `gorm:"default:0"`
	AvgFillPrice   float64   `gorm:"type:decimal(20,4)"`
	Status         string    `gorm:"type:varchar(20);index;not null"`
	Version        int       `gorm:"not null;default:0"`
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`
}

func (OrderReadModelDB) TableName() string { return "order_read_models" }

// ─────────────────────────────────────────────────────────────────────────────
// toDBReadModel / toDomainReadModel helpers
// ─────────────────────────────────────────────────────────────────────────────

func toDBReadModel(rm *domain.OrderReadModel) *OrderReadModelDB {
	return &OrderReadModelDB{
		ID:             rm.ID,
		UserID:         rm.UserID,
		AccountID:      rm.AccountID,
		Symbol:         rm.Symbol,
		Side:           string(rm.Side),
		OrderType:      string(rm.OrderType),
		Quantity:       rm.Quantity,
		Price:          rm.Price,
		FilledQuantity: rm.FilledQuantity,
		AvgFillPrice:   rm.AvgFillPrice,
		Status:         string(rm.Status),
		Version:        rm.Version,
		CreatedAt:      rm.CreatedAt,
		UpdatedAt:      rm.UpdatedAt,
	}
}

func toDomainReadModel(m *OrderReadModelDB) *domain.OrderReadModel {
	return &domain.OrderReadModel{
		ID:             m.ID,
		UserID:         m.UserID,
		AccountID:      m.AccountID,
		Symbol:         m.Symbol,
		Side:           domain.Side(m.Side),
		OrderType:      domain.OrderType(m.OrderType),
		Quantity:       m.Quantity,
		Price:          m.Price,
		FilledQuantity: m.FilledQuantity,
		AvgFillPrice:   m.AvgFillPrice,
		Status:         domain.Status(m.Status),
		Version:        m.Version,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// eventStore — implements EventStore (append-only log)
// ─────────────────────────────────────────────────────────────────────────────

type eventStore struct {
	db *gorm.DB
}

// NewEventStore creates a new Postgres-backed order event store.
func NewEventStore(db *gorm.DB) EventStore {
	return &eventStore{db: db}
}

// AppendEvents saves new events with optimistic concurrency guard.
func (s *eventStore) AppendEvents(ctx context.Context, aggregateID string, expectedVersion int, events []EventDescriptor) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Verify current max version equals expectedVersion (optimistic concurrency)
		var currentVersion int
		tx.Model(&OrderEventModel{}).
			Where("aggregate_id = ?", aggregateID).
			Select("COALESCE(MAX(version), 0)").
			Scan(&currentVersion)

		if currentVersion != expectedVersion {
			return fmt.Errorf("optimistic concurrency conflict: expected version %d but current is %d", expectedVersion, currentVersion)
		}

		for _, ed := range events {
			row := &OrderEventModel{
				ID:          uuid.New().String(),
				AggregateID: ed.AggregateID,
				EventType:   string(ed.EventType),
				Payload:     ed.Payload,
				Version:     ed.Version,
				OccurredAt:  ed.OccurredAt,
			}
			if err := tx.Create(row).Error; err != nil {
				return fmt.Errorf("failed to append event %s: %w", ed.EventType, err)
			}
		}
		return nil
	})
}

// AppendEventsWithHook runs the same optimistic-concurrency insert as
// AppendEvents but also calls afterFn(tx) inside the same DB transaction.
// This allows callers (e.g., EventSourcingService) to insert outbox rows
// atomically alongside domain events.
func (s *eventStore) AppendEventsWithHook(ctx context.Context, aggregateID string, expectedVersion int, events []EventDescriptor, afterFn func(tx *gorm.DB) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var currentVersion int
		tx.Model(&OrderEventModel{}).
			Where("aggregate_id = ?", aggregateID).
			Select("COALESCE(MAX(version), 0)").
			Scan(&currentVersion)

		if currentVersion != expectedVersion {
			return fmt.Errorf("optimistic concurrency conflict: expected version %d but current is %d", expectedVersion, currentVersion)
		}

		for _, ed := range events {
			row := &OrderEventModel{
				ID:          uuid.New().String(),
				AggregateID: ed.AggregateID,
				EventType:   string(ed.EventType),
				Payload:     ed.Payload,
				Version:     ed.Version,
				OccurredAt:  ed.OccurredAt,
			}
			if err := tx.Create(row).Error; err != nil {
				return fmt.Errorf("failed to append event %s: %w", ed.EventType, err)
			}
		}

		if afterFn != nil {
			return afterFn(tx)
		}
		return nil
	})
}

// LoadEvents fetches all events for an aggregate ordered by version ASC.
func (s *eventStore) LoadEvents(ctx context.Context, aggregateID string) ([]EventDescriptor, error) {
	var rows []OrderEventModel
	err := s.db.WithContext(ctx).
		Where("aggregate_id = ?", aggregateID).
		Order("version ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toDescriptors(rows), nil
}

// LoadAllDescriptors fetches every event ordered by (aggregate_id, version) ASC.
func (s *eventStore) LoadAllDescriptors(ctx context.Context) ([]EventDescriptor, error) {
	var rows []OrderEventModel
	err := s.db.WithContext(ctx).
		Order("aggregate_id ASC, version ASC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("LoadAllDescriptors: %w", err)
	}
	return toDescriptors(rows), nil
}

func toDescriptors(rows []OrderEventModel) []EventDescriptor {
	descs := make([]EventDescriptor, 0, len(rows))
	for _, r := range rows {
		descs = append(descs, EventDescriptor{
			ID:          r.ID,
			AggregateID: r.AggregateID,
			EventType:   domain.EventType(r.EventType),
			Payload:     r.Payload,
			Version:     r.Version,
			OccurredAt:  r.OccurredAt,
		})
	}
	return descs
}

// ─────────────────────────────────────────────────────────────────────────────
// readModelRepo — implements domain.ReadModelRepository
// ─────────────────────────────────────────────────────────────────────────────

type readModelRepo struct {
	db *gorm.DB
}

// NewReadModelRepository creates a new Postgres-backed order read model repository.
func NewReadModelRepository(db *gorm.DB) domain.ReadModelRepository {
	return &readModelRepo{db: db}
}

// Upsert creates or replaces the read model for a given order (idempotent).
func (r *readModelRepo) Upsert(ctx context.Context, rm *domain.OrderReadModel) error {
	model := toDBReadModel(rm)
	return r.db.WithContext(ctx).
		Where("id = ?", model.ID).
		Assign(model).
		FirstOrCreate(model).Error
}

// GetByID returns the read model for a single order.
func (r *readModelRepo) GetByID(ctx context.Context, id string) (*domain.OrderReadModel, error) {
	var m OrderReadModelDB
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		return nil, fmt.Errorf("order read model not found: %w", err)
	}
	return toDomainReadModel(&m), nil
}

// ListByUserID retrieves all orders for a user ordered by created_at DESC.
func (r *readModelRepo) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.OrderReadModel, error) {
	var models []*OrderReadModelDB
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toDomainReadModels(models), nil
}

// ListByUserIDAndFilter filters by optional symbol and status.
func (r *readModelRepo) ListByUserIDAndFilter(ctx context.Context, userID, symbol, status string, limit, offset int) ([]*domain.OrderReadModel, error) {
	q := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if symbol != "" {
		q = q.Where("symbol = ?", symbol)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}

	var models []*OrderReadModelDB
	err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toDomainReadModels(models), nil
}

func toDomainReadModels(models []*OrderReadModelDB) []*domain.OrderReadModel {
	result := make([]*domain.OrderReadModel, 0, len(models))
	for _, m := range models {
		result = append(result, toDomainReadModel(m))
	}
	return result
}
