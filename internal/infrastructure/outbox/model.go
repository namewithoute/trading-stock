package outbox

import (
	"time"

	"gorm.io/gorm"
)

// OutboxEventModel is the GORM persistence model for the transactional outbox table.
// Rows are inserted atomically inside the same DB transaction as the domain event,
// then asynchronously published to Kafka by the OutboxRelay.
//
// Table: outbox_events
type OutboxEventModel struct {
	ID          string         `gorm:"column:id;primaryKey;type:uuid"`
	Topic       string         `gorm:"column:topic;not null;index"`
	MessageKey  string         `gorm:"column:message_key;not null"`
	Payload     []byte         `gorm:"column:payload;not null"`
	CreatedAt   time.Time      `gorm:"column:created_at;not null;index"`
	ProcessedAt *time.Time     `gorm:"column:processed_at;index"` // NULL = not yet published
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (OutboxEventModel) TableName() string { return "outbox_events" }
