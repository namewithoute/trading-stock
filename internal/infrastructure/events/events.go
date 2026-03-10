// Package events defines shared Kafka message types used across multiple
// infrastructure packages, avoiding import cycles.
package events

import (
	"time"

	pkgdecimal "trading-stock/pkg/decimal"
)

// KafkaTopicTradesExecuted is the topic that carries TradeExecutedMessage.
const KafkaTopicTradesExecuted = "trades.executed"

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
