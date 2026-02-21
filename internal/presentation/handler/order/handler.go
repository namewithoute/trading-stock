package order

import (
	"net/http"
	orderUC "trading-stock/internal/application/order"

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
	accountID := "" // Will let usecase handle getting the primary account natively
	createdOrder, err := h.orderUseCase.CreateOrder(
		c.Request().Context(),
		userID.(string),
		accountID,
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

	return c.JSON(http.StatusCreated, Order{
		OrderID:   createdOrder.ID,
		Symbol:    createdOrder.Symbol,
		Side:      string(createdOrder.Side),
		Type:      string(createdOrder.Type),
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

	var dtoList []Order
	for _, o := range orders {
		dtoList = append(dtoList, Order{
			OrderID:        o.ID,
			Symbol:         o.Symbol,
			Side:           string(o.Side),
			Type:           string(o.Type),
			Quantity:       o.Quantity,
			FilledQuantity: o.FilledQuantity,
			Price:          o.Price,
			Status:         string(o.Status),
			CreatedAt:      o.CreatedAt,
		})
	}

	if dtoList == nil {
		dtoList = make([]Order, 0)
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
	userID := c.Get("user_id")

	// TODO: Implement get order detail logic
	// 1. Get order ID from URL param
	// 2. Get user ID from context
	// 3. Verify order belongs to user
	// 4. Fetch order details from database
	// 5. Include trade history for this order
	// 6. Return full order info

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Get order detail - TODO: implement",
		"user_id": userID,
		"data": map[string]interface{}{
			"order_id":        orderID,
			"symbol":          "VNM",
			"side":            "buy",
			"type":            "limit",
			"quantity":        100,
			"filled_quantity": 50,
			"price":           85000,
			"average_price":   84500,
			"status":          "partial_filled",
			"trades": []map[string]interface{}{
				{
					"trade_id":  "trd_001",
					"quantity":  30,
					"price":     84000,
					"timestamp": "2024-01-01T10:01:00Z",
				},
				{
					"trade_id":  "trd_002",
					"quantity":  20,
					"price":     85000,
					"timestamp": "2024-01-01T10:02:00Z",
				},
			},
		},
	})
}

// CancelOrder cancels an order (protected)
// DELETE /api/v1/orders/:id
func (h *OrderHandler) CancelOrder(c echo.Context) error {
	orderID := c.Param("id")
	userID := c.Get("user_id")

	// TODO: Implement cancel order logic
	// 1. Get order ID and user ID
	// 2. Verify order belongs to user
	// 3. Check if order can be cancelled (not filled/cancelled)
	// 4. Send cancel request to matching engine
	// 5. Update order status to "cancelled"
	// 6. Release frozen balance
	// 7. Return success

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":  "Order cancelled successfully",
		"user_id":  userID,
		"order_id": orderID,
	})
}

// UpdateOrder updates an order (protected)
// PUT /api/v1/orders/:id
func (h *OrderHandler) UpdateOrder(c echo.Context) error {
	orderID := c.Param("id")
	userID := c.Get("user_id")

	// TODO: Implement update order logic
	// 1. Get order ID and user ID
	// 2. Verify order belongs to user
	// 3. Parse request body (new_price, new_quantity)
	// 4. Validate new values
	// 5. Cancel old order
	// 6. Create new order with updated values
	// 7. Return updated order info
	// Note: In stock trading, usually cancel + create new order instead of update

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Order updated successfully",
		"user_id": userID,
		"data": map[string]interface{}{
			"order_id":  orderID,
			"new_price": 86000,
			"status":    "pending",
		},
	})
}
