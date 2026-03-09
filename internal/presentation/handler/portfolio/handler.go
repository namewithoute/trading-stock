package portfolio

import (
	"net/http"
	portfolioUC "trading-stock/internal/application/portfolio"
	"trading-stock/pkg/response"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// PortfolioHandler handles portfolio management endpoints
type PortfolioHandler struct {
	portfolioUseCase portfolioUC.UseCase
	logger           *zap.Logger
}

// NewPortfolioHandler creates a new portfolio handler
func NewPortfolioHandler(portfolioUseCase portfolioUC.UseCase, logger *zap.Logger) *PortfolioHandler {
	return &PortfolioHandler{
		portfolioUseCase: portfolioUseCase,
		logger:           logger,
	}
}

// GetOverview gets portfolio overview (protected)
// GET /api/v1/portfolio
func (h *PortfolioHandler) GetOverview(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
	}

	positions, err := h.portfolioUseCase.GetOverview(c.Request().Context(), userID)
	if err != nil {
		h.logger.Error("GetOverview failed", zap.Error(err), zap.String("user_id", userID))
		return response.Error(c, http.StatusInternalServerError, "Failed to get portfolio overview", err.Error())
	}

	totalValue, err := h.portfolioUseCase.GetTotalValue(c.Request().Context(), userID)
	if err != nil {
		h.logger.Warn("GetTotalValue failed", zap.Error(err))
	}

	totalPnL, err := h.portfolioUseCase.GetTotalUnrealizedPnL(c.Request().Context(), userID)
	if err != nil {
		h.logger.Warn("GetTotalUnrealizedPnL failed", zap.Error(err))
	}

	dtos := make([]PositionDTO, 0, len(positions))
	for _, p := range positions {
		dtos = append(dtos, toPositionDTO(p))
	}

	return response.Success(c, http.StatusOK, "Portfolio overview retrieved", PortfolioOverviewResponse{
		UserID:        userID,
		TotalValue:    totalValue,
		TotalPnL:      totalPnL,
		PositionCount: len(positions),
		Positions:     dtos,
	})
}

// ListPositions lists all positions (protected)
// GET /api/v1/portfolio/positions
func (h *PortfolioHandler) ListPositions(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
	}

	positions, err := h.portfolioUseCase.GetOverview(c.Request().Context(), userID)
	if err != nil {
		h.logger.Error("ListPositions failed", zap.Error(err), zap.String("user_id", userID))
		return response.Error(c, http.StatusInternalServerError, "Failed to list positions", err.Error())
	}

	dtos := make([]PositionDTO, 0, len(positions))
	for _, p := range positions {
		dtos = append(dtos, toPositionDTO(p))
	}

	return response.Success(c, http.StatusOK, "Positions retrieved", dtos)
}

// GetPosition gets position by symbol (protected)
// GET /api/v1/portfolio/positions/:symbol
func (h *PortfolioHandler) GetPosition(c echo.Context) error {
	symbol := c.Param("symbol")
	accountID := c.QueryParam("account_id")

	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
	}

	if symbol == "" {
		return response.Error(c, http.StatusBadRequest, "Symbol is required", "symbol_empty")
	}

	// accountID is optional; fall back to empty string for GetPositionBySymbol
	pos, err := h.portfolioUseCase.GetPositionBySymbol(c.Request().Context(), accountID, symbol)
	if err != nil {
		h.logger.Error("GetPosition failed", zap.Error(err), zap.String("symbol", symbol))
		return response.Error(c, http.StatusInternalServerError, "Failed to get position", err.Error())
	}

	if pos == nil {
		return response.Error(c, http.StatusNotFound, "Position not found", "not_found")
	}

	return response.Success(c, http.StatusOK, "Position retrieved", toPositionDTO(pos))
}

// GetPerformance gets portfolio performance summary (protected)
// GET /api/v1/portfolio/performance
func (h *PortfolioHandler) GetPerformance(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
	}

	totalValue, err := h.portfolioUseCase.GetTotalValue(c.Request().Context(), userID)
	if err != nil {
		h.logger.Error("GetPerformance/GetTotalValue failed", zap.Error(err))
		return response.Error(c, http.StatusInternalServerError, "Failed to get performance", err.Error())
	}

	totalPnL, err := h.portfolioUseCase.GetTotalUnrealizedPnL(c.Request().Context(), userID)
	if err != nil {
		h.logger.Warn("GetTotalUnrealizedPnL failed", zap.Error(err))
	}

	var pnlPercent float64
	costBasis := totalValue - totalPnL
	if costBasis > 0 {
		pnlPercent = (totalPnL / costBasis) * 100
	}

	return response.Success(c, http.StatusOK, "Portfolio performance retrieved", PortfolioPerformanceResponse{
		UserID:             userID,
		TotalValue:         totalValue,
		TotalUnrealizedPnL: totalPnL,
		PnLPercent:         pnlPercent,
	})
}
