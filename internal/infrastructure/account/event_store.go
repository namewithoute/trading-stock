package account

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	domain "trading-stock/internal/domain/account"

	"github.com/cockroachdb/apd/v3"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────────────────────────────────
// GORM Models for the two new tables
// ─────────────────────────────────────────────────────────────────────────────

// AccountEventModel maps to `account_events` – the immutable append-only log.
type AccountEventModel struct {
	ID          string    `gorm:"primaryKey;type:uuid"`
	AggregateID string    `gorm:"type:uuid;index;not null"`
	EventType   string    `gorm:"type:varchar(64);not null"`
	Payload     []byte    `gorm:"type:jsonb;not null"`
	Version     int       `gorm:"not null"`
	OccurredAt  time.Time `gorm:"not null"`
}

func (AccountEventModel) TableName() string { return "account_events" }

// AccountReadModel maps to `account_read_models` – the denormalised query table.
type AccountReadModelDB struct {
	ID          string      `gorm:"primaryKey;type:uuid"`
	UserID      string      `gorm:"type:uuid;index;not null"`
	AccountType string      `gorm:"type:varchar(20);not null"`
	Balance     apd.Decimal `gorm:"type:decimal(20,2);not null;default:0"`
	BuyingPower apd.Decimal `gorm:"type:decimal(20,2);not null;default:0"`
	Currency    string      `gorm:"type:varchar(3);not null;default:'USD'"`
	Status      string      `gorm:"type:varchar(20);not null"`
	Version     int         `gorm:"not null;default:0"`
	UpdatedAt   time.Time   `gorm:"not null"`
}

func (AccountReadModelDB) TableName() string { return "account_read_models" }

// ─────────────────────────────────────────────────────────────────────────────
// eventStore – implements domain.EventStore
// ─────────────────────────────────────────────────────────────────────────────

type eventStore struct {
	db *gorm.DB
}

// NewEventStore creates a new Postgres-backed event store.
func NewEventStore(db *gorm.DB) EventStore {
	return &eventStore{db: db}
}

// AppendEvents saves new events with optimistic concurrency guard.
func (s *eventStore) AppendEvents(ctx context.Context, aggregateID string, expectedVersion int, events []EventDescriptor) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check the current max version for this aggregate
		var currentVersion int
		tx.Model(&AccountEventModel{}).
			Where("aggregate_id = ?", aggregateID).
			Select("COALESCE(MAX(version), 0)").
			Scan(&currentVersion)

		if currentVersion != expectedVersion {
			return fmt.Errorf("optimistic concurrency conflict: expected version %d but got %d", expectedVersion, currentVersion)
		}

		// Append each event
		for _, ed := range events {
			row := &AccountEventModel{
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
func (s *eventStore) AppendEventsWithHook(ctx context.Context, aggregateID string, expectedVersion int, events []EventDescriptor, afterFn func(tx *gorm.DB) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var currentVersion int
		tx.Model(&AccountEventModel{}).
			Where("aggregate_id = ?", aggregateID).
			Select("COALESCE(MAX(version), 0)").
			Scan(&currentVersion)

		if currentVersion != expectedVersion {
			return fmt.Errorf("optimistic concurrency conflict: expected version %d but got %d", expectedVersion, currentVersion)
		}

		for _, ed := range events {
			row := &AccountEventModel{
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
	var rows []AccountEventModel
	err := s.db.WithContext(ctx).
		Where("aggregate_id = ?", aggregateID).
		Order("version ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toDescriptors(rows), nil
}

// LoadEventsFrom fetches events from a specific version onward.
func (s *eventStore) LoadEventsFrom(ctx context.Context, aggregateID string, fromVersion int) ([]EventDescriptor, error) {
	var rows []AccountEventModel
	err := s.db.WithContext(ctx).
		Where("aggregate_id = ? AND version > ?", aggregateID, fromVersion).
		Order("version ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toDescriptors(rows), nil
}

// LoadAllDescriptors fetches every event in the store ordered by (aggregate_id, version) ASC.
// Called once on startup by the Projector to rebuild all read models before switching to
// live Kafka streaming. In high-volume systems, replace with a paginated cursor.
func (s *eventStore) LoadAllDescriptors(ctx context.Context) ([]EventDescriptor, error) {
	var rows []AccountEventModel
	err := s.db.WithContext(ctx).
		Order("aggregate_id ASC, version ASC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("LoadAllDescriptors: %w", err)
	}
	return toDescriptors(rows), nil
}

func toDescriptors(rows []AccountEventModel) []EventDescriptor {
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
// readModelRepo – implements domain.ReadModelRepository
// ─────────────────────────────────────────────────────────────────────────────

type readModelRepo struct {
	db *gorm.DB
}

// NewReadModelRepository creates a Postgres-backed read model repo.
func NewReadModelRepository(db *gorm.DB) domain.ReadModelRepository {
	return &readModelRepo{db: db}
}

// Upsert inserts or updates the read model for an account.
func (r *readModelRepo) Upsert(ctx context.Context, rm *domain.AccountReadModel) error {
	row := &AccountReadModelDB{
		ID:          rm.ID,
		UserID:      rm.UserID,
		AccountType: string(rm.AccountType),
		Balance:     rm.Balance,
		BuyingPower: rm.BuyingPower,
		Currency:    rm.Currency,
		Status:      string(rm.Status),
		Version:     rm.Version,
		UpdatedAt:   time.Now().UTC(),
	}

	return r.db.WithContext(ctx).
		Save(row). // Upsert: insert if new, update if PK exists
		Error
}

// GetByID retrieves a single read model.
func (r *readModelRepo) GetByID(ctx context.Context, id string) (*domain.AccountReadModel, error) {
	var row AccountReadModelDB
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, err
	}
	return toReadModelDomain(&row), nil
}

// GetByUserID retrieves all read models for a user.
func (r *readModelRepo) GetByUserID(ctx context.Context, userID string) ([]*domain.AccountReadModel, error) {
	var rows []AccountReadModelDB
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("updated_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]*domain.AccountReadModel, 0, len(rows))
	for i := range rows {
		result = append(result, toReadModelDomain(&rows[i]))
	}
	return result, nil
}

func toReadModelDomain(row *AccountReadModelDB) *domain.AccountReadModel {
	return &domain.AccountReadModel{
		ID:          row.ID,
		UserID:      row.UserID,
		AccountType: domain.AccountType(row.AccountType),
		Balance:     row.Balance,
		BuyingPower: row.BuyingPower,
		Currency:    row.Currency,
		Status:      domain.Status(row.Status),
		Version:     row.Version,
		UpdatedAt:   row.UpdatedAt,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Helper – Deserialise raw JSON payload back to domain.DomainEvent
// ─────────────────────────────────────────────────────────────────────────────

// DeserialiseEvent reconstructs a typed DomainEvent from an EventDescriptor.
func DeserialiseEvent(ed EventDescriptor) (domain.DomainEvent, error) {
	switch ed.EventType {
	case domain.EventAccountCreated:
		var e domain.AccountCreatedEvent
		if err := json.Unmarshal(ed.Payload, &e); err != nil {
			return nil, err
		}
		return e, nil

	case domain.EventMoneyDeposited:
		var e domain.MoneyDepositedEvent
		if err := json.Unmarshal(ed.Payload, &e); err != nil {
			return nil, err
		}
		return e, nil

	case domain.EventMoneyWithdrawn:
		var e domain.MoneyWithdrawnEvent
		if err := json.Unmarshal(ed.Payload, &e); err != nil {
			return nil, err
		}
		return e, nil

	case domain.EventFundsReserved:
		var e domain.FundsReservedEvent
		if err := json.Unmarshal(ed.Payload, &e); err != nil {
			return nil, err
		}
		return e, nil

	case domain.EventFundsReleased:
		var e domain.FundsReleasedEvent
		if err := json.Unmarshal(ed.Payload, &e); err != nil {
			return nil, err
		}
		return e, nil

	case domain.EventStatusChanged:
		var e domain.StatusChangedEvent
		if err := json.Unmarshal(ed.Payload, &e); err != nil {
			return nil, err
		}
		return e, nil

	case domain.EventTradeSettled:
		var e domain.TradeSettledEvent
		if err := json.Unmarshal(ed.Payload, &e); err != nil {
			return nil, err
		}
		return e, nil

	default:
		return nil, fmt.Errorf("unknown event type: %s", ed.EventType)
	}
}
