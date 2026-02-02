package engine

import (
	"sync"
	"trading-stock/internal/domain"
)

type MatchingEngine struct {
	orderBook *domain.OrderBook
	mu        sync.Mutex
}

func NewMatchingEngine(orderBook *domain.OrderBook) *MatchingEngine {
	return &MatchingEngine{
		orderBook: orderBook,
	}
}

func (e *MatchingEngine) MatchOrder(order *domain.Order) ([]domain.Trade, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	trades := e.orderBook.ProcessOrder(order)
	return trades, nil
}
