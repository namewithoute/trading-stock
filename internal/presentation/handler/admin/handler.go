package admin

import (
	"net/http"
	"strconv"
	adminUC "trading-stock/internal/application/admin"
	"trading-stock/internal/domain/user"
	"trading-stock/pkg/response"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// AdminHandler handles admin endpoints
type AdminHandler struct {
	adminUseCase adminUC.UseCase
	logger       *zap.Logger
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(adminUseCase adminUC.UseCase, logger *zap.Logger) *AdminHandler {
	return &AdminHandler{
		adminUseCase: adminUseCase,
		logger:       logger,
	}
}

// ListUsers lists all users (admin only)
// GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c echo.Context) error {
	limit := 20
	offset := 0

	if l := c.QueryParam("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	if o := c.QueryParam("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	users, err := h.adminUseCase.ListUsers(c.Request().Context(), limit, offset)
	if err != nil {
		h.logger.Error("ListUsers failed", zap.Error(err))
		return response.Error(c, http.StatusInternalServerError, "Failed to list users", err.Error())
	}

	dtos := make([]UserAdminDTO, 0, len(users))
	for _, u := range users {
		dtos = append(dtos, toUserAdminDTO(u))
	}

	return response.Success(c, http.StatusOK, "Users retrieved", map[string]interface{}{
		"data": dtos,
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"count":  len(dtos),
		},
	})
}

// ApproveKYC approves or rejects a user's KYC (admin only)
// PUT /api/v1/admin/users/:id/kyc
func (h *AdminHandler) ApproveKYC(c echo.Context) error {
	targetUserID := c.Param("id")
	if targetUserID == "" {
		return response.Error(c, http.StatusBadRequest, "User ID is required", "user_id_empty")
	}

	var req ApproveKYCRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
	}

	kycStatus := user.KYCStatus(req.Status)
	if !kycStatus.IsValid() {
		return response.Error(c, http.StatusBadRequest, "Invalid KYC status, must be PENDING, APPROVED, or REJECTED", "invalid_status")
	}

	if err := h.adminUseCase.ApproveKYC(c.Request().Context(), targetUserID, kycStatus); err != nil {
		h.logger.Error("ApproveKYC failed", zap.Error(err), zap.String("target_user_id", targetUserID))
		return response.Error(c, http.StatusInternalServerError, "Failed to update KYC status", err.Error())
	}

	return response.Success(c, http.StatusOK, "KYC status updated successfully", map[string]interface{}{
		"user_id":    targetUserID,
		"kyc_status": string(kycStatus),
	})
}

// ListAllOrders lists all orders in the system (admin only)
// GET /api/v1/admin/orders
func (h *AdminHandler) ListAllOrders(c echo.Context) error {
	limit := 20
	offset := 0

	if l := c.QueryParam("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	if o := c.QueryParam("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	orders, total, err := h.adminUseCase.ListAllOrders(c.Request().Context(), limit, offset)
	if err != nil {
		h.logger.Error("ListAllOrders failed", zap.Error(err))
		return response.Error(c, http.StatusInternalServerError, "Failed to list orders", err.Error())
	}

	dtos := make([]OrderAdminDTO, 0, len(orders))
	for _, o := range orders {
		dtos = append(dtos, toOrderAdminDTO(o))
	}

	return response.Success(c, http.StatusOK, "Orders retrieved", map[string]interface{}{
		"data": dtos,
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  total,
			"count":  len(dtos),
		},
	})
}

// GetSystemStats gets system statistics (admin only)
// GET /api/v1/admin/stats
func (h *AdminHandler) GetSystemStats(c echo.Context) error {
	stats, err := h.adminUseCase.GetSystemStats(c.Request().Context())
	if err != nil {
		h.logger.Error("GetSystemStats failed", zap.Error(err))
		return response.Error(c, http.StatusInternalServerError, "Failed to get system stats", err.Error())
	}

	return response.Success(c, http.StatusOK, "System statistics retrieved", stats)
}
