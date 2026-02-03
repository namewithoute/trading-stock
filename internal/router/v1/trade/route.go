package trade

import (
	"trading-stock/internal/handler/trade"

	"github.com/labstack/echo/v4"
)

type TradeRouter struct {
	handler *trade.TradeHandler
}

func NewTradeRouter(handler *trade.TradeHandler) *TradeRouter {
	return &TradeRouter{handler: handler}
}

// Public routes - không cần auth
func (r *TradeRouter) RegisterPublicRoutes(g *echo.Group) {
	trades := g.Group("/trades")

	// Ví dụ: Lịch sử giao dịch công khai của thị trường (không phải của user)
	trades.GET("/market/:symbol", r.handler.GetMarketTrades) // GET /api/v1/trades/market/VNM
}

// Protected routes - cần auth
func (r *TradeRouter) RegisterRoutes(g *echo.Group) {
	trades := g.Group("/trades")

	trades.GET("", r.handler.ListTrades)         // GET /api/v1/trades (của user)
	trades.GET("/:id", r.handler.GetTradeDetail) // GET /api/v1/trades/:id
}
