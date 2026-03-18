// Package events defines shared Kafka message types used across multiple
// infrastructure packages, avoiding import cycles.
package events

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	domainOrder "trading-stock/internal/domain/order"
	pkgdecimal "trading-stock/pkg/decimal"
)

// KafkaTopicTradesExecuted is the topic that carries TradeExecutedMessage.
const KafkaTopicTradesExecuted = "trades.executed"

// KafkaTopicOrdersMarketExpired is the topic for market orders whose unfilled
// remainder should be expired (Immediate-or-Cancel semantics).
const KafkaTopicOrdersMarketExpired = "orders.market_expired"

const KafkaTopicOrdersUpdated = "trading.orders.updated"

// TradeExecutedMessage is the payload published to trades.executed.
// Produced by: matching service.
// Consumed by: order fill consumer, account trade consumer, market data consumer.
type TradeExecutedMessage struct {
	EventID     string             `json:"event_id"`
	TradeID     string             `json:"trade_id"`
	Symbol      string             `json:"symbol"`
	Price       pkgdecimal.Decimal `json:"price"`
	Quantity    int                `json:"quantity"`
	BuyOrderID  string             `json:"buy_order_id"`
	SellOrderID string             `json:"sell_order_id"`
	BuyerID     string             `json:"buyer_id"`
	SellerID    string             `json:"seller_id"`
	OccurredAt  time.Time          `json:"occurred_at"`
}

// MarketOrderExpiredMessage is published when a market order finishes matching
// with unfilled remainder (including zero fills). The consumer loads the aggregate,
// waits until all fills have been applied, then expires the order.
type MarketOrderExpiredMessage struct {
	EventID        string    `json:"event_id"`
	OrderID        string    `json:"order_id"`
	FilledQuantity int       `json:"filled_quantity"` // total qty filled by the engine
	OccurredAt     time.Time `json:"occurred_at"`
}

// OrderUpdatedMessage is emitted by the matching engine whenever an order's
// matching state changes (e.g. partial fill, filled, expired).
//
// This message enables downstream domains to react without importing the engine
// package directly.
type OrderUpdatedMessage struct {
	EventID        string               `json:"event_id"`
	OrderID        string               `json:"order_id"`
	UserID         string               `json:"user_id"`
	AccountID      string               `json:"account_id"`
	Symbol         string               `json:"symbol"`
	Side           domainOrder.Side     `json:"side"`
	OrderType      domainOrder.OrderType `json:"order_type"`
	Status         domainOrder.Status   `json:"status"`
	Quantity       int                  `json:"quantity"`
	FilledQuantity int                  `json:"filled_quantity"`
	Price          pkgdecimal.Decimal   `json:"price"`
	AvgFillPrice   pkgdecimal.Decimal   `json:"avg_fill_price"`
	OccurredAt     time.Time            `json:"occurred_at"`
}

// DecodeKafkaPayload unmarshals both plain JSON messages and Debezium/Kafka
// Connect schema-wrapped messages into out.
func DecodeKafkaPayload(raw []byte, out interface{}) error {
	// 1) Plain JSON payload (the happy path for native producers)
	if err := json.Unmarshal(raw, out); err == nil {
		return nil
	}

	// 2) Connect wrapper payload: {"schema":...,"payload":...}
	var wrapper map[string]json.RawMessage
	if err := json.Unmarshal(raw, &wrapper); err == nil {
		if payloadRaw, hasPayload := wrapper["payload"]; hasPayload {
			if decodeErr := decodeFlexibleJSON(payloadRaw, out); decodeErr == nil {
				return nil
			}
		}
	}

	// 3) Entire record is a JSON string (common when converter emits quoted payload)
	if err := decodeFlexibleJSON(raw, out); err == nil {
		return nil
	}

	return fmt.Errorf("invalid JSON payload")
}

func decodeFlexibleJSON(raw []byte, out interface{}) error {
	// A) Raw JSON object/array payload
	if err := json.Unmarshal(raw, out); err == nil {
		return nil
	}

	// B) String payload that may contain JSON or base64(JSON)
	var str string
	if err := json.Unmarshal(raw, &str); err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(str), out); err == nil {
		return nil
	}

	decoded, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return err
	}

	return json.Unmarshal(decoded, out)
}
