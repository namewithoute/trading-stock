package order

import (
	"trading-stock/internal/presentation/handler/order"

	"github.com/labstack/echo/v4"
)

type OrderRouter struct {
	handler *order.OrderHandler
}

func NewOrderRouter(handler *order.OrderHandler) *OrderRouter {
	return &OrderRouter{handler: handler}
}

// Protected routes - cần auth
func (r *OrderRouter) RegisterRoutes(g *echo.Group) {
	orders := g.Group("/orders")

	orders.POST("", r.handler.CreateOrder)       // POST /api/v1/orders
	orders.GET("", r.handler.ListOrders)         // GET /api/v1/orders
	orders.GET("/:id", r.handler.GetOrderDetail) // GET /api/v1/orders/:id
	orders.DELETE("/:id", r.handler.CancelOrder) // DELETE /api/v1/orders/:id
	orders.PUT("/:id", r.handler.UpdateOrder)    // PUT /api/v1/orders/:id
}

func (r *OrderRouter) RegisterPublicRoutes(g *echo.Group) {

}
