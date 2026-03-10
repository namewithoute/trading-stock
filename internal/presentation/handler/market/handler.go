package market

import (
	"net/http"
	"time"

	marketUC "trading-stock/internal/application/market"
	pkgdecimal "trading-stock/pkg/decimal"
	"trading-stock/pkg/response"

	"github.com/cockroachdb/apd/v3"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var handlerDecCtx = apd.BaseContext.WithPrecision(19)

// MarketHandler handles market data endpoints.
type MarketHandler struct {
	marketUseCase marketUC.UseCase
	logger        *zap.Logger
}

func NewMarketHandler(marketUseCase marketUC.UseCase, logger *zap.Logger) *MarketHandler {
	return &MarketHandler{marketUseCase: marketUseCase, logger: logger}
}

// GetTrendingStocks GET /api/v1/market/trending
func (h *MarketHandler) GetTrendingStocks(c echo.Context) error {
	limitStr := c.QueryParam("limit")
	limit := 10
	if limitStr != "" {
		if n, err := parseInt(limitStr); err == nil && n > 0 {
			limit = n
		}
	}

	results, err := h.marketUseCase.GetTrendingStocks(c.Request().Context(), limit)
	if err != nil {
		h.logger.Error("GetTrendingStocks failed", zap.Error(err))
		return response.Error(c, http.StatusInternalServerError, "Failed to get trending stocks", err.Error())
	}

	dtos := make([]TrendingStockDTO, 0, len(results))
	for _, r := range results {
		dto := TrendingStockDTO{
			Symbol:   r.Stock.Symbol,
			Name:     r.Stock.Name,
			Exchange: r.Stock.Exchange,
		}
		if r.LatestPrice != nil {
			dto.Price = pkgdecimal.From(r.LatestPrice.Price)
			dto.Bid = pkgdecimal.From(r.LatestPrice.Bid)
			dto.Ask = pkgdecimal.From(r.LatestPrice.Ask)
			dto.Volume = r.LatestPrice.Volume
		}
		dtos = append(dtos, dto)
	}

	return response.Success(c, http.StatusOK, "Trending stocks retrieved", dtos)
}

// ListStocks GET /api/v1/market/stocks
func (h *MarketHandler) ListStocks(c echo.Context) error {
	var req ListStocksRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Invalid query parameters", err.Error())
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	offset := (req.Page - 1) * req.Limit

	results, err := h.marketUseCase.ListStocks(c.Request().Context(), req.Exchange, req.Sector, req.Search, req.Limit, offset)
	if err != nil {
		h.logger.Error("ListStocks failed", zap.Error(err))
		return response.Error(c, http.StatusInternalServerError, "Failed to list stocks", err.Error())
	}

	dtos := make([]StockDTO, 0, len(results))
	for _, r := range results {
		dto := StockDTO{
			Symbol:     r.Stock.Symbol,
			Name:       r.Stock.Name,
			Exchange:   r.Stock.Exchange,
			Sector:     r.Stock.Sector,
			Industry:   r.Stock.Industry,
			IsActive:   r.Stock.IsActive,
			IsTradable: r.Stock.IsTradable,
		}
		if r.LatestPrice != nil {
			dto.Price = pkgdecimal.From(r.LatestPrice.Price)
			dto.Bid = pkgdecimal.From(r.LatestPrice.Bid)
			dto.Ask = pkgdecimal.From(r.LatestPrice.Ask)
			dto.Volume = r.LatestPrice.Volume
			dto.PriceAt = r.LatestPrice.Timestamp
		}
		dtos = append(dtos, dto)
	}

	return response.Success(c, http.StatusOK, "Stocks retrieved", ListStocksResponse{
		Stocks: dtos,
		Pagination: Pagination{
			Page:  req.Page,
			Limit: req.Limit,
			Total: len(dtos),
		},
	})
}

// GetStockDetail GET /api/v1/market/stocks/:symbol
func (h *MarketHandler) GetStockDetail(c echo.Context) error {
	symbol := c.Param("symbol")

	r, err := h.marketUseCase.GetStockDetail(c.Request().Context(), symbol)
	if err != nil {
		h.logger.Error("GetStockDetail failed", zap.Error(err), zap.String("symbol", symbol))
		return response.Error(c, http.StatusNotFound, "Stock not found", err.Error())
	}

	resp := StockDetailResponse{
		Symbol:     r.Stock.Symbol,
		Name:       r.Stock.Name,
		Exchange:   r.Stock.Exchange,
		Sector:     r.Stock.Sector,
		Industry:   r.Stock.Industry,
		IsActive:   r.Stock.IsActive,
		IsTradable: r.Stock.IsTradable,
		CreatedAt:  r.Stock.CreatedAt,
	}
	if r.LatestPrice != nil {
		resp.Price = pkgdecimal.From(r.LatestPrice.Price)
		resp.Bid = pkgdecimal.From(r.LatestPrice.Bid)
		resp.Ask = pkgdecimal.From(r.LatestPrice.Ask)
		var spread apd.Decimal
		_, _ = handlerDecCtx.Sub(&spread, &r.LatestPrice.Ask, &r.LatestPrice.Bid)
		resp.Spread = pkgdecimal.From(spread)
		resp.Volume = r.LatestPrice.Volume
		resp.PriceAt = r.LatestPrice.Timestamp
	}

	return response.Success(c, http.StatusOK, "Stock detail retrieved", resp)
}

// GetCurrentPrice GET /api/v1/market/stocks/:symbol/price
func (h *MarketHandler) GetCurrentPrice(c echo.Context) error {
	symbol := c.Param("symbol")

	p, err := h.marketUseCase.GetLatestPrice(c.Request().Context(), symbol)
	if err != nil {
		h.logger.Error("GetCurrentPrice failed", zap.Error(err), zap.String("symbol", symbol))
		return response.Error(c, http.StatusNotFound, "Price not found", err.Error())
	}

	return response.Success(c, http.StatusOK, "Current price retrieved", PriceResponse{
		Symbol:    symbol,
		Price:     pkgdecimal.From(p.Price),
		Bid:       pkgdecimal.From(p.Bid),
		Ask:       pkgdecimal.From(p.Ask),
		Spread:    func() pkgdecimal.Decimal { var s apd.Decimal; _, _ = handlerDecCtx.Sub(&s, &p.Ask, &p.Bid); return pkgdecimal.From(s) }(),
		Volume:    p.Volume,
		Timestamp: p.Timestamp,
	})
}

// GetPriceHistory GET /api/v1/market/stocks/:symbol/price/history
func (h *MarketHandler) GetPriceHistory(c echo.Context) error {
	symbol := c.Param("symbol")

	var req PriceHistoryRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Invalid query parameters", err.Error())
	}

	now := time.Now().UTC()
	from := now.Add(-24 * time.Hour)
	to := now

	if req.From != "" {
		if t, err := time.Parse(time.RFC3339, req.From); err == nil {
			from = t
		}
	}
	if req.To != "" {
		if t, err := time.Parse(time.RFC3339, req.To); err == nil {
			to = t
		}
	}

	prices, err := h.marketUseCase.GetPriceHistory(c.Request().Context(), symbol, from, to)
	if err != nil {
		h.logger.Error("GetPriceHistory failed", zap.Error(err), zap.String("symbol", symbol))
		return response.Error(c, http.StatusInternalServerError, "Failed to get price history", err.Error())
	}

	dtos := make([]PriceResponse, 0, len(prices))
	for _, p := range prices {
		dtos = append(dtos, PriceResponse{
			Symbol:    symbol,
			Price:     pkgdecimal.From(p.Price),
			Bid:       pkgdecimal.From(p.Bid),
			Ask:       pkgdecimal.From(p.Ask),
			Spread:    func() pkgdecimal.Decimal { var s apd.Decimal; _, _ = handlerDecCtx.Sub(&s, &p.Ask, &p.Bid); return pkgdecimal.From(s) }(),
			Volume:    p.Volume,
			Timestamp: p.Timestamp,
		})
	}

	return response.Success(c, http.StatusOK, "Price history retrieved", map[string]interface{}{
		"symbol": symbol,
		"from":   from,
		"to":     to,
		"count":  len(dtos),
		"prices": dtos,
	})
}

// GetCandles GET /api/v1/market/stocks/:symbol/candles
func (h *MarketHandler) GetCandles(c echo.Context) error {
	symbol := c.Param("symbol")

	var req GetCandlesRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Invalid query parameters", err.Error())
	}
	if req.Interval == "" {
		req.Interval = "1d"
	}

	now := time.Now().UTC()
	from := now.AddDate(0, -1, 0) // default: last 30 days
	to := now

	if req.From != "" {
		if t, err := time.Parse(time.RFC3339, req.From); err == nil {
			from = t
		}
	}
	if req.To != "" {
		if t, err := time.Parse(time.RFC3339, req.To); err == nil {
			to = t
		}
	}

	candles, err := h.marketUseCase.GetCandles(c.Request().Context(), symbol, req.Interval, from, to)
	if err != nil {
		h.logger.Error("GetCandles failed", zap.Error(err), zap.String("symbol", symbol))
		return response.Error(c, http.StatusInternalServerError, "Failed to get candles", err.Error())
	}

	dtos := make([]CandleDTO, 0, len(candles))
	for _, k := range candles {
		dtos = append(dtos, CandleDTO{
			Timestamp: k.Timestamp,
			Open:      pkgdecimal.From(k.Open),
			High:      pkgdecimal.From(k.High),
			Low:       pkgdecimal.From(k.Low),
			Close:     pkgdecimal.From(k.Close),
			Volume:    k.Volume,
		})
	}

	return response.Success(c, http.StatusOK, "Candles retrieved", GetCandlesResponse{
		Symbol:   symbol,
		Interval: req.Interval,
		From:     from,
		To:       to,
		Count:    len(dtos),
		Candles:  dtos,
	})
}

// GetOrderBook GET /api/v1/market/stocks/:symbol/orderbook
// Returns live order-book depth sourced from the matching engine.
// NOTE: The engine exposes its order book via the in-process EngineService; a
// future iteration will inject that service here. For now we return an empty
// book so the endpoint contract is stable.
func (h *MarketHandler) GetOrderBook(c echo.Context) error {
	symbol := c.Param("symbol")
	return response.Success(c, http.StatusOK, "Order book retrieved", map[string]interface{}{
		"symbol":    symbol,
		"bids":      []interface{}{},
		"asks":      []interface{}{},
		"timestamp": time.Now().UTC(),
		"note":      "Live order book will be available once the engine service is wired to this handler",
	})
}

// GetPremiumAnalysis GET /api/v1/market/premium/analysis/:symbol  (protected)
// Computes basic technical indicators from the last 30 days of daily candles.
func (h *MarketHandler) GetPremiumAnalysis(c echo.Context) error {
	symbol := c.Param("symbol")
	userID := c.Get("user_id").(string)

	ctx := c.Request().Context()
	now := time.Now().UTC()
	from := now.AddDate(0, -1, 0)

	candles, err := h.marketUseCase.GetCandles(ctx, symbol, "1d", from, now)
	if err != nil || len(candles) == 0 {
		return response.Success(c, http.StatusOK, "Analysis unavailable (insufficient data)", map[string]interface{}{
			"symbol":  symbol,
			"user_id": userID,
		})
	}

	// Simple Moving Average (SMA-14)
	n := len(candles)
	smaWindow := 14
	if n < smaWindow {
		smaWindow = n
	}
	var sum apd.Decimal
	for i := n - smaWindow; i < n; i++ {
		_, _ = handlerDecCtx.Add(&sum, &sum, &candles[i].Close)
	}
	var sma14 apd.Decimal
	_, _ = handlerDecCtx.Quo(&sma14, &sum, apd.New(int64(smaWindow), 0))

	// Last close & simple momentum
	last := candles[n-1]
	var momentum apd.Decimal
	if n >= 2 {
		prevClose := candles[n-2].Close
		var diff apd.Decimal
		_, _ = handlerDecCtx.Sub(&diff, &last.Close, &prevClose)
		if prevClose.Sign() != 0 {
			var pct apd.Decimal
			_, _ = handlerDecCtx.Quo(&pct, &diff, &prevClose)
			_, _ = handlerDecCtx.Mul(&momentum, &pct, apd.New(100, 0))
		}
	}

	// Naive recommendation
	recommendation := "hold"
	var buyThreshold, sellThreshold apd.Decimal
	_, _ = handlerDecCtx.Mul(&buyThreshold, &sma14, apd.New(102, -2)) // sma14 * 1.02
	_, _ = handlerDecCtx.Mul(&sellThreshold, &sma14, apd.New(98, -2)) // sma14 * 0.98
	if last.Close.Cmp(&buyThreshold) > 0 {
		recommendation = "buy"
	} else if last.Close.Cmp(&sellThreshold) < 0 {
		recommendation = "sell"
	}

	return response.Success(c, http.StatusOK, "Premium analysis retrieved", map[string]interface{}{
		"symbol":         symbol,
		"user_id":        userID,
		"recommendation": recommendation,
		"last_close":     pkgdecimal.From(last.Close),
		"sma_14":         pkgdecimal.From(sma14),
		"momentum_pct":   pkgdecimal.From(momentum),
		"candles_used":   n,
		"as_of":          last.Timestamp,
	})
}

// GetWatchlist GET /api/v1/market/watchlist  (protected)
// Watchlist persistence is not yet implemented (requires a watchlist table).
func (h *MarketHandler) GetWatchlist(c echo.Context) error {
	userID := c.Get("user_id").(string)
	return response.Success(c, http.StatusOK, "Watchlist retrieved", map[string]interface{}{
		"user_id": userID,
		"stocks":  []interface{}{},
		"note":    "Watchlist persistence is not yet implemented",
	})
}

// AddToWatchlist POST /api/v1/market/watchlist  (protected)
func (h *MarketHandler) AddToWatchlist(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req AddWatchlistRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
	}
	if req.Symbol == "" {
		return response.Error(c, http.StatusBadRequest, "symbol is required", "symbol is required")
	}

	// Validate the symbol exists
	if _, err := h.marketUseCase.GetStockDetail(c.Request().Context(), req.Symbol); err != nil {
		return response.Error(c, http.StatusNotFound, "Symbol not found", err.Error())
	}

	h.logger.Info("Watchlist add requested (persistence not yet implemented)",
		zap.String("userID", userID),
		zap.String("symbol", req.Symbol),
	)
	return response.Success(c, http.StatusCreated, "Symbol added to watchlist (persistence coming soon)", map[string]string{
		"user_id": userID,
		"symbol":  req.Symbol,
	})
}

// parseInt parses a string as a base-10 integer.
func parseInt(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, echo.ErrBadRequest
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
