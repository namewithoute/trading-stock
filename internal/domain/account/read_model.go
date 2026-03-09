package account

import (
	"time"

	"github.com/cockroachdb/apd/v3"
)

// AccountReadModel is the denormalised, query-optimised view of an account.
// It lives in the `account_read_models` table and is safely reconstructed
// from the EventStore by the Projector.
//
// In CQRS architecture, this model is used STRICTLY for Read operations (Queries).
// Writes (Commands) never touch this model, they operate on the AccountAggregate.
type AccountReadModel struct {
	ID          string      `json:"id"`
	UserID      string      `json:"user_id"`
	AccountType AccountType `json:"account_type"`
	Balance     apd.Decimal `json:"balance"`
	BuyingPower apd.Decimal `json:"buying_power"`
	Currency    string      `json:"currency"`
	Status      Status      `json:"status"`

	// Version matches the latest event version applied to this read model.
	// It is used by the Idempotent Consumer (Projector) to drop duplicate
	// or out-of-order Kafka events.
	Version   int       `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
}
