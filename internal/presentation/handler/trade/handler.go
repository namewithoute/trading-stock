package trade

import (
	"net/http"
	"strconv"
	executionUC "trading-stock/internal/application/execution"
	"trading-stock/pkg/response"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// TradeHandler handles trade history endpoints
type TradeHandler struct {
	tradeUseCase executionUC.UseCase
	logger       *zap.Logger
}

// NewTradeHandler creates a new trade handler
func NewTradeHandler(tradeUseCase executionUC.UseCase, logger *zap.Logger) *TradeHandler {
	return &TradeHandler{
		tradeUseCase: tradeUseCase,
		logger:       logger,
	}
}

// GetMarketTrades gets market trades for a symbol (public)
// GET /api/v1/trades/market/:symbol
func (h *TradeHandler) GetMarketTrades(c echo.Context) error {
	symbol := c.Param("symbol")
	if symbol == "" {
		return response.Error(c, http.StatusBadRequest, "Symbol is required", "symbol_empty")
	}

	limit := 20
	if l := c.QueryParam("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	trades, err := h.tradeUseCase.GetMarketTrades(c.Request().Context(), symbol, limit)
	if err != nil {
		h.logger.Error("GetMarketTrades failed", zap.Error(err), zap.String("symbol", symbol))
		return response.Error(c, http.StatusInternalServerError, "Failed to get market trades", err.Error())
	}

	dtos := make([]TradeDTO, 0, len(trades))
	for _, t := range trades {
		dtos = append(dtos, toTradeDTO(t))
	}

	return response.Success(c, http.StatusOK, "Market trades retrieved", dtos)
}

// ListTrades lists user's trade history (protected)
// GET /api/v1/trades
func (h *TradeHandler) ListTrades(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
	}

	trades, err := h.tradeUseCase.ListTrades(c.Request().Context(), userID)
	if err != nil {
		h.logger.Error("ListTrades failed", zap.Error(err), zap.String("user_id", userID))
		return response.Error(c, http.StatusInternalServerError, "Failed to list trades", err.Error())
	}

	dtos := make([]TradeDTO, 0, len(trades))
	for _, t := range trades {
		dtos = append(dtos, toTradeDTO(t))
	}

	return response.Success(c, http.StatusOK, "Trade history retrieved", dtos)
}

// GetTradeDetail gets trade details (protected)
// GET /api/v1/trades/:id
func (h *TradeHandler) GetTradeDetail(c echo.Context) error {
	tradeID := c.Param("id")
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
	}

	if tradeID == "" {
		return response.Error(c, http.StatusBadRequest, "Trade ID is required", "trade_id_empty")
	}

	trade, err := h.tradeUseCase.GetTradeDetail(c.Request().Context(), tradeID)
	if err != nil {
		h.logger.Error("GetTradeDetail failed", zap.Error(err), zap.String("trade_id", tradeID))
		return response.Error(c, http.StatusNotFound, "Trade not found", err.Error())
	}

	// Verify the trade belongs to the requesting user.
	if trade.BuyerID != userID && trade.SellerID != userID {
		return response.Error(c, http.StatusForbidden, "Access denied", "you do not own this trade")
	}

	return response.Success(c, http.StatusOK, "Trade detail retrieved", toTradeDTO(trade))
}
