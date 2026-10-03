package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/middleware"
	"github.com/pareshvernekar/homecooked/internal/models"
)

// OrderService port for HTTP handlers.
type OrderService interface {
	Create(ctx context.Context, tenantID string, req *models.OrderCreateRequest) (*models.CustomerOrder, error)
	List(ctx context.Context, tenantID string) ([]*models.CustomerOrder, error)
	Get(ctx context.Context, tenantID, id string) (*models.OrderDetail, error)
	Update(ctx context.Context, tenantID, id string, req *models.OrderUpdateRequest) (*models.OrderDetail, error)
	AddLine(ctx context.Context, tenantID, orderID string, req *models.OrderLineCreateRequest) (*models.OrderItem, error)
	UpdateLine(ctx context.Context, tenantID, orderID, itemID string, req *models.OrderLineUpdateRequest) (*models.OrderItem, error)
	RemoveLine(ctx context.Context, tenantID, orderID, itemID string) error
	RecordPayment(ctx context.Context, tenantID, orderID string, req *models.PaymentCreateRequest) (*models.OrderPayment, error)
	ListPayments(ctx context.Context, tenantID, orderID string) ([]models.OrderPayment, error)
}

// OrderHandler serves /api/v1/orders.
// REQORDER001–REQORDER005, REQOLINE001–REQOLINE004, REQPAY001–REQPAY004
type OrderHandler struct {
	orders OrderService
	Logger *logger.Logger
}

// NewOrderHandler constructs an order handler.
func NewOrderHandler(orders OrderService, l *logger.Logger) *OrderHandler {
	return &OrderHandler{orders: orders, Logger: l}
}

func badRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"success": false, "error_code": "INVALID_REQUEST", "message": err.Error(), "timestamp": time.Now().UTC()})
}

// CreateOrder POST /orders
// REQORDER001
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.OrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	tenantID := c.GetString(middleware.TenantIDKey)
	o, err := h.orders.Create(c.Request.Context(), tenantID, &req)
	if err != nil {
		writeServiceError(c, err, "Failed to create order")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Order created successfully", "data": o})
}

// ListOrders GET /orders
// REQORDER002
func (h *OrderHandler) ListOrders(c *gin.Context) {
	tenantID := c.GetString(middleware.TenantIDKey)
	orders, err := h.orders.List(c.Request.Context(), tenantID)
	if err != nil {
		writeServiceError(c, err, "Failed to list orders")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Orders retrieved successfully", "data": orders})
}

// GetOrder GET /orders/:id
// REQORDER002
func (h *OrderHandler) GetOrder(c *gin.Context) {
	tenantID := c.GetString(middleware.TenantIDKey)
	o, err := h.orders.Get(c.Request.Context(), tenantID, c.Param("id"))
	if err != nil {
		writeServiceError(c, err, "Failed to get order")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Order retrieved successfully", "data": o})
}

// UpdateOrder PATCH /orders/:id
// REQORDER003, REQORDER004
func (h *OrderHandler) UpdateOrder(c *gin.Context) {
	var req models.OrderUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	tenantID := c.GetString(middleware.TenantIDKey)
	o, err := h.orders.Update(c.Request.Context(), tenantID, c.Param("id"), &req)
	if err != nil {
		writeServiceError(c, err, "Failed to update order")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Order updated successfully", "data": o})
}

// AddOrderItem POST /orders/:id/items
// REQOLINE001
func (h *OrderHandler) AddOrderItem(c *gin.Context) {
	var req models.OrderLineCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	tenantID := c.GetString(middleware.TenantIDKey)
	item, err := h.orders.AddLine(c.Request.Context(), tenantID, c.Param("id"), &req)
	if err != nil {
		writeServiceError(c, err, "Failed to add order item")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Order item created successfully", "data": item})
}

// UpdateOrderItem PATCH /orders/:id/items/:itemId
// REQOLINE002, REQOLINE004
func (h *OrderHandler) UpdateOrderItem(c *gin.Context) {
	var req models.OrderLineUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	tenantID := c.GetString(middleware.TenantIDKey)
	item, err := h.orders.UpdateLine(c.Request.Context(), tenantID, c.Param("id"), c.Param("itemId"), &req)
	if err != nil {
		writeServiceError(c, err, "Failed to update order item")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Order item updated successfully", "data": item})
}

// DeleteOrderItem DELETE /orders/:id/items/:itemId
// REQOLINE004
func (h *OrderHandler) DeleteOrderItem(c *gin.Context) {
	tenantID := c.GetString(middleware.TenantIDKey)
	if err := h.orders.RemoveLine(c.Request.Context(), tenantID, c.Param("id"), c.Param("itemId")); err != nil {
		writeServiceError(c, err, "Failed to remove order item")
		return
	}
	c.Status(http.StatusNoContent)
}

// CreatePayment POST /orders/:id/payments
// REQPAY001
func (h *OrderHandler) CreatePayment(c *gin.Context) {
	var req models.PaymentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	tenantID := c.GetString(middleware.TenantIDKey)
	p, err := h.orders.RecordPayment(c.Request.Context(), tenantID, c.Param("id"), &req)
	if err != nil {
		writeServiceError(c, err, "Failed to record payment")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Payment recorded successfully", "data": p})
}

// ListPayments GET /orders/:id/payments
// REQPAY004
func (h *OrderHandler) ListPayments(c *gin.Context) {
	tenantID := c.GetString(middleware.TenantIDKey)
	payments, err := h.orders.ListPayments(c.Request.Context(), tenantID, c.Param("id"))
	if err != nil {
		writeServiceError(c, err, "Failed to list payments")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Payments retrieved successfully", "data": payments})
}
