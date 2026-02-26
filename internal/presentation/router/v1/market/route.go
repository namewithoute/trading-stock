package market

import (
	"trading-stock/internal/presentation/handler/market"

	"github.com/labstack/echo/v4"
)

type MarketRouter struct {
	handler *market.MarketHandler
}

func NewMarketRouter(handler *market.MarketHandler) *MarketRouter {
	return &MarketRouter{handler: handler}
}

// Public routes - không cần auth (market data ai cũng xem được)
func (r *MarketRouter) RegisterPublicRoutes(g *echo.Group) {
	market := g.Group("/market")
	market.GET("/trending", r.handler.GetTrendingStocks) // GET /api/v1/market/stocks/:symbol/orderbook

	stocks := market.Group("/stocks")

	stocks.GET("", r.handler.ListStocks)                            // GET /api/v1/market/stocks
	stocks.GET("/:symbol", r.handler.GetStockDetail)                // GET /api/v1/market/stocks/:symbol
	stocks.GET("/:symbol/price", r.handler.GetCurrentPrice)         // GET /api/v1/market/stocks/:symbol/price
	stocks.GET("/:symbol/price/history", r.handler.GetPriceHistory) // GET /api/v1/market/stocks/:symbol/price/history
	stocks.GET("/:symbol/candles", r.handler.GetCandles)            // GET /api/v1/market/stocks/:symbol/candles
	stocks.GET("/:symbol/orderbook", r.handler.GetOrderBook)        // GET /api/v1/market/stocks/:symbol/orderbook

}

// Protected routes - cần auth
func (r *MarketRouter) RegisterRoutes(g *echo.Group) {
	market := g.Group("/market")

	// Ví dụ: Premium market data chỉ cho user đã auth
	market.GET("/premium/analysis/:symbol", r.handler.GetPremiumAnalysis)
	market.GET("/watchlist", r.handler.GetWatchlist) // Watchlist cá nhân
	market.POST("/watchlist", r.handler.AddToWatchlist)
}
