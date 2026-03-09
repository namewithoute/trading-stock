package risk

import (
	"net/http"
	riskUC "trading-stock/internal/application/risk"
	"trading-stock/pkg/response"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// RiskHandler handles risk management endpoints.
type RiskHandler struct {
	riskUseCase riskUC.UseCase
	logger      *zap.Logger
}

// NewRiskHandler creates a new RiskHandler.
func NewRiskHandler(riskUseCase riskUC.UseCase, logger *zap.Logger) *RiskHandler {
	return &RiskHandler{
		riskUseCase: riskUseCase,
		logger:      logger,
	}
}

// GetRiskMetrics returns current risk metrics for an account.
// GET /api/v1/risk/metrics/:account_id
func (h *RiskHandler) GetRiskMetrics(c echo.Context) error {
	accountID := c.Param("account_id")
	if accountID == "" {
		return response.Error(c, http.StatusBadRequest, "account_id is required", "account_id_empty")
	}

	metrics, err := h.riskUseCase.GetRiskMetrics(c.Request().Context(), accountID)
	if err != nil {
		h.logger.Error("GetRiskMetrics failed", zap.Error(err), zap.String("account_id", accountID))
		return response.Error(c, http.StatusInternalServerError, "Failed to get risk metrics", err.Error())
	}

	return response.Success(c, http.StatusOK, "Risk metrics retrieved", RiskMetricsResponse{
		AccountID:          metrics.AccountID,
		UserID:             metrics.UserID,
		TotalExposure:      metrics.TotalExposure,
		LongExposure:       metrics.LongExposure,
		ShortExposure:      metrics.ShortExposure,
		PositionsCount:     metrics.PositionsCount,
		LargestPosition:    metrics.LargestPosition,
		ConcentrationRatio: metrics.ConcentrationRatio,
		DailyPnL:           metrics.DailyPnL,
		WeeklyPnL:          metrics.WeeklyPnL,
		MonthlyPnL:         metrics.MonthlyPnL,
		DailyOrdersCount:   metrics.DailyOrdersCount,
		CurrentLeverage:    metrics.CurrentLeverage,
		RiskScore:          metrics.RiskScore,
		IsHighRisk:         metrics.IsHighRisk(),
		IsMediumRisk:       metrics.IsMediumRisk(),
		LastCalculatedAt:   metrics.LastCalculatedAt,
	})
}

// GetActiveAlerts returns active risk alerts for an account.
// GET /api/v1/risk/alerts/:account_id
func (h *RiskHandler) GetActiveAlerts(c echo.Context) error {
	accountID := c.Param("account_id")
	if accountID == "" {
		return response.Error(c, http.StatusBadRequest, "account_id is required", "account_id_empty")
	}

	alerts, err := h.riskUseCase.GetActiveAlerts(c.Request().Context(), accountID)
	if err != nil {
		h.logger.Error("GetActiveAlerts failed", zap.Error(err), zap.String("account_id", accountID))
		return response.Error(c, http.StatusInternalServerError, "Failed to get risk alerts", err.Error())
	}

	dtos := make([]RiskAlertResponse, 0, len(alerts))
	for _, a := range alerts {
		dtos = append(dtos, RiskAlertResponse{
			ID:         a.ID,
			AccountID:  a.AccountID,
			UserID:     a.UserID,
			AlertType:  string(a.AlertType),
			Severity:   string(a.Severity),
			Message:    a.Message,
			Status:     string(a.Status),
			CreatedAt:  a.CreatedAt,
			ResolvedAt: a.ResolvedAt,
		})
	}

	return response.Success(c, http.StatusOK, "Risk alerts retrieved", dtos)
}

// CheckOrderRisk checks whether a pending order passes risk controls.
// POST /api/v1/risk/check
func (h *RiskHandler) CheckOrderRisk(c echo.Context) error {
	var req CheckOrderRiskRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Invalid request payload", err.Error())
	}
	if req.AccountID == "" || req.Symbol == "" || req.Amount <= 0 {
		return response.Error(c, http.StatusBadRequest, "account_id, symbol and amount are required", "invalid_params")
	}

	passed, err := h.riskUseCase.CheckOrderRisk(c.Request().Context(), req.AccountID, req.Symbol, req.Amount)
	if err != nil {
		h.logger.Error("CheckOrderRisk failed", zap.Error(err))
		return response.Error(c, http.StatusInternalServerError, "Risk check failed", err.Error())
	}

	msg := "Order passed risk controls"
	if !passed {
		msg = "Order rejected by risk controls"
	}

	return response.Success(c, http.StatusOK, msg, map[string]interface{}{
		"passed":     passed,
		"account_id": req.AccountID,
		"symbol":     req.Symbol,
		"amount":     req.Amount,
	})
}
