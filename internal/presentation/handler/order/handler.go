package order

import (
	"net/http"
	orderUC "trading-stock/internal/application/order"
	"trading-stock/pkg/response"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// OrderHandler handles order management endpoints
type OrderHandler struct {
	orderUseCase orderUC.UseCase
	logger       *zap.Logger
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(orderUseCase orderUC.UseCase, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{
		orderUseCase: orderUseCase,
		logger:       logger,
	}
}

// CreateOrder creates a new order (protected)
// POST /api/v1/orders
func (h *OrderHandler) CreateOrder(c echo.Context) error {
	userID := c.Get("user_id")

	// TODO: Implement create order logic
	// 1. Get user ID from context
	var orderReq CreateOrderRequest
	if err := c.Bind(&orderReq); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if orderReq.Symbol == "" || orderReq.Price == 0 || orderReq.Quantity == 0 || orderReq.Side == "" || orderReq.Type == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid request body",
			"error":   "Invalid request body",
		})
	}

	// Call UseCase
	createdOrder, err := h.orderUseCase.CreateOrder(
		c.Request().Context(),
		userID.(string),
		orderReq.AccountID,
		orderReq.Symbol,
		orderReq.Side,
		orderReq.Type,
		orderReq.Price,
		int(orderReq.Quantity),
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to create order",
			"error":   err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, OrderDTO{
		OrderID:   createdOrder.ID,
		Symbol:    createdOrder.Symbol,
		Side:      string(createdOrder.Side),
		Type:      string(createdOrder.OrderType),
		Quantity:  createdOrder.Quantity,
		Price:     createdOrder.Price,
		Status:    string(createdOrder.Status),
		CreatedAt: createdOrder.CreatedAt,
	})
}

// ListOrders lists all orders of current user (protected)
// GET /api/v1/orders
func (h *OrderHandler) ListOrders(c echo.Context) error {
	userID := c.Get("user_id")

	var listOrdersRequest ListOrdersRequest
	if err := c.Bind(&listOrdersRequest); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	page := listOrdersRequest.Page
	if page <= 0 {
		page = 1
	}
	limit := listOrdersRequest.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit

	orders, err := h.orderUseCase.ListOrders(
		c.Request().Context(),
		userID.(string),
		listOrdersRequest.Symbol,
		listOrdersRequest.Status,
		limit,
		offset,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to list orders",
			"error":   err.Error(),
		})
	}

	var dtoList []OrderDTO
	for _, o := range orders {
		dtoList = append(dtoList, OrderDTO{
			OrderID:        o.ID,
			Symbol:         o.Symbol,
			Side:           string(o.Side),
			Type:           string(o.OrderType),
			Quantity:       o.Quantity,
			FilledQuantity: o.FilledQuantity,
			Price:          o.Price,
			Status:         string(o.Status),
			CreatedAt:      o.CreatedAt,
		})
	}

	if dtoList == nil {
		dtoList = make([]OrderDTO, 0)
	}

	return c.JSON(http.StatusOK, ListOrdersResponse{
		Orders: dtoList,
		Pagination: Pagination{
			Page:  page,
			Limit: limit,
			Total: len(orders), // TODO: Add a real total count query instead of len(orders)
		},
	})
}

// GetOrderDetail gets order details (protected)
// GET /api/v1/orders/:id
func (h *OrderHandler) GetOrderDetail(c echo.Context) error {
	orderID := c.Param("id")
	userID := c.Get("user_id").(string)

	// 1. Fetch order from repository via usecase
	o, err := h.orderUseCase.GetOrder(c.Request().Context(), orderID)
	if err != nil {
		h.logger.Error("Failed to get order", zap.Error(err), zap.String("orderID", orderID))
		return response.Error(c, http.StatusNotFound, "Order not found", err.Error())
	}

	// 2. Verify the order belongs to the requesting user
	if o.UserID != userID {
		return response.Error(c, http.StatusForbidden, "Access denied", "you do not own this order")
	}

	return response.Success(c, http.StatusOK, "Order retrieved", GetOrderDetailResponse{
		OrderID:        o.ID,
		AccountID:      o.AccountID,
		Symbol:         o.Symbol,
		Side:           string(o.Side),
		Type:           string(o.OrderType),
		Quantity:       o.Quantity,
		FilledQuantity: o.FilledQuantity,
		Price:          o.Price,
		AvgFillPrice:   o.AvgFillPrice,
		Status:         string(o.Status),
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
	})
}

// CancelOrder cancels an order (protected)
// DELETE /api/v1/orders/:id
func (h *OrderHandler) CancelOrder(c echo.Context) error {
	orderID := c.Param("id")
	userID := c.Get("user_id").(string)

	// 1. Fetch order to verify ownership before cancelling
	o, err := h.orderUseCase.GetOrder(c.Request().Context(), orderID)
	if err != nil {
		h.logger.Error("Order not found for cancel", zap.Error(err), zap.String("orderID", orderID))
		return response.Error(c, http.StatusNotFound, "Order not found", err.Error())
	}

	// 2. Verify ownership
	if o.UserID != userID {
		return response.Error(c, http.StatusForbidden, "Access denied", "you do not own this order")
	}

	// 3. Cancel via usecase (also releases reserved BUY funds internally)
	if err := h.orderUseCase.CancelOrder(c.Request().Context(), orderID); err != nil {
		h.logger.Error("Failed to cancel order", zap.Error(err), zap.String("orderID", orderID))
		return response.Error(c, http.StatusBadRequest, "Failed to cancel order", err.Error())
	}

	h.logger.Info("Order cancelled", zap.String("orderID", orderID), zap.String("userID", userID))
	return response.Success(c, http.StatusOK, "Order cancelled successfully", map[string]string{
		"order_id": orderID,
	})
}

// UpdateOrder updates an order (protected)
// PUT /api/v1/orders/:id
// In stock trading, modification is implemented as cancel-then-recreate:
// the old order is cancelled (releasing any reserved funds) and a new order is
// placed with the updated price / quantity.
func (h *OrderHandler) UpdateOrder(c echo.Context) error {
	orderID := c.Param("id")
	userID := c.Get("user_id").(string)

	// 1. Parse and validate request body
	var req UpdateOrderRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
	}
	if req.Price <= 0 || req.Quantity <= 0 {
		return response.Error(c, http.StatusBadRequest, "Invalid request body", "price and quantity must be greater than 0")
	}

	// 2. Delegate to usecase (ownership check + cancel + recreate happens there)
	updatedOrder, err := h.orderUseCase.UpdateOrder(c.Request().Context(), userID, orderID, req.Price, req.Quantity)
	if err != nil {
		h.logger.Error("Failed to update order",
			zap.Error(err),
			zap.String("orderID", orderID),
			zap.String("userID", userID),
		)
		return response.Error(c, http.StatusBadRequest, "Failed to update order", err.Error())
	}

	h.logger.Info("Order updated", zap.String("newOrderID", updatedOrder.ID), zap.String("userID", userID))
	return response.Success(c, http.StatusOK, "Order updated successfully", GetOrderDetailResponse{
		OrderID:        updatedOrder.ID,
		AccountID:      updatedOrder.AccountID,
		Symbol:         updatedOrder.Symbol,
		Side:           string(updatedOrder.Side),
		Type:           string(updatedOrder.OrderType),
		Quantity:       updatedOrder.Quantity,
		FilledQuantity: updatedOrder.FilledQuantity,
		Price:          updatedOrder.Price,
		AvgFillPrice:   updatedOrder.AvgFillPrice,
		Status:         string(updatedOrder.Status),
		CreatedAt:      updatedOrder.CreatedAt,
		UpdatedAt:      updatedOrder.UpdatedAt,
	})
}
