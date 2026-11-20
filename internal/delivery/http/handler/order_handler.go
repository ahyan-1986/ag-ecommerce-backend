package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/delivery/http/middleware"
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/ahyan-1986/ag-ecommerce-backend/pkg/pdf"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

type OrderHandler struct {
	orderSvc       domain.OrderService
	paymentLogRepo domain.PaymentLogRepository
	webhookSecret  string
}

func NewOrderHandler(orderSvc domain.OrderService, paymentLogRepo domain.PaymentLogRepository, webhookSecret string) *OrderHandler {
	return &OrderHandler{
		orderSvc:       orderSvc,
		paymentLogRepo: paymentLogRepo,
		webhookSecret:  webhookSecret,
	}
}

// ── Cart ─────────────────────────────────────────────────────────────────────

func (h *OrderHandler) GetCart(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	items, err := h.orderSvc.GetCart(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *OrderHandler) AddToCart(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req domain.CartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	items, err := h.orderSvc.AddToCart(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *OrderHandler) UpdateCartItem(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item ID"})
		return
	}
	var req struct {
		Quantity int `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	items, err := h.orderSvc.UpdateCartItem(userID, itemID, req.Quantity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *OrderHandler) RemoveCartItem(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item ID"})
		return
	}
	if err := h.orderSvc.RemoveCartItem(userID, itemID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "item removed from cart"})
}

// ── Checkout & Orders ─────────────────────────────────────────────────────────

func (h *OrderHandler) GuestCheckout(c *gin.Context) {
	var req domain.GuestCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order, err := h.orderSvc.GuestCheckout(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"order": order})
}

func (h *OrderHandler) Checkout(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req domain.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order, paymentURL, err := h.orderSvc.Checkout(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"order":      order,
		"paymentUrl": paymentURL,
	})
}

func (h *OrderHandler) GetMyOrders(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	orders, total, err := h.orderSvc.ListUserOrders(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": orders, "total": total, "page": page, "pageSize": pageSize})
}

func (h *OrderHandler) GetMyOrder(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}
	order, err := h.orderSvc.GetOrder(id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) CancelMyOrder(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}
	if err := h.orderSvc.CancelOrder(id, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "order cancelled"})
}

func (h *OrderHandler) CaptureMyOrderPayment(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	orderNo := c.Param("orderNo")
	if orderNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order number"})
		return
	}
	// sessionId is passed from the Stripe success URL {CHECKOUT_SESSION_ID} template
	var body struct {
		SessionID string `json:"sessionId"`
	}
	c.ShouldBindJSON(&body) // body is optional; ignore parse errors

	order, err := h.orderSvc.CaptureStripePayment(userID, orderNo, body.SessionID)
	if err != nil {
		msg := err.Error()
		// payment_pending means Stripe hasn't settled yet — tell frontend to retry
		if strings.HasPrefix(msg, "payment_pending:") {
			c.JSON(http.StatusAccepted, gin.H{"error": msg, "pending": true})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"order": order, "message": "payment verified"})
}

func (h *OrderHandler) GetMyOrderByNo(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	orderNo := c.Param("orderNo")
	order, err := h.orderSvc.GetOrderByNo(userID, orderNo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

// Admin endpoints

func (h *OrderHandler) ListOrders(c *gin.Context) {
	status := domain.OrderStatus(c.Query("status"))
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	orders, total, err := h.orderSvc.ListAllOrders(status, search, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": orders, "total": total, "page": page, "pageSize": pageSize})
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}
	order, err := h.orderSvc.GetOrder(id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) UpdateFulfillment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}
	var req domain.UpdateFulfillmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order, err := h.orderSvc.UpdateFulfillment(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.orderSvc.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// Stripe webhook — verifies signature and logs every inbound event.
func (h *OrderHandler) StripeWebhook(c *gin.Context) {
	const maxBytes = int64(65536)
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBytes))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	sigHeader := c.GetHeader("Stripe-Signature")

	// Verify Stripe signature when webhook secret is configured.
	var event stripe.Event
	if h.webhookSecret != "" {
		event, err = webhook.ConstructEvent(body, sigHeader, h.webhookSecret)
		if err != nil {
			_ = h.paymentLogRepo.Create(&domain.PaymentLog{
				Gateway:      "stripe",
				EventType:    "unknown",
				Status:       domain.PaymentLogFailed,
				ErrorMessage: "signature verification failed: " + err.Error(),
				IPAddress:    c.ClientIP(),
				RawPayload:   string(body),
			})
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature"})
			return
		}
	} else {
		// Dev/test fallback — no signature check
		if err := json.Unmarshal(body, &event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}
	}

	// Extract useful fields from checkout.session.completed
	var sessionID, orderNo, currency string
	var amountCents int64
	if event.Type == "checkout.session.completed" {
		var sess stripe.CheckoutSession
		if jsonErr := json.Unmarshal(event.Data.Raw, &sess); jsonErr == nil {
			sessionID = sess.ID
			amountCents = sess.AmountTotal
			currency = strings.ToUpper(string(sess.Currency))
			if sess.Metadata != nil {
				orderNo = sess.Metadata["order_no"]
			}
		}
	}

	log := &domain.PaymentLog{
		Gateway:      "stripe",
		EventType:    string(event.Type),
		GatewayTxnID: sessionID,
		OrderNo:      orderNo,
		AmountCents:  amountCents,
		Currency:     currency,
		Status:       domain.PaymentLogReceived,
		IPAddress:    c.ClientIP(),
		RawPayload:   string(body),
	}
	_ = h.paymentLogRepo.Create(log)

	switch event.Type {
	case "checkout.session.completed":
		if processErr := h.orderSvc.ConfirmStripePayment(sessionID); processErr != nil {
			log.Status = domain.PaymentLogFailed
			log.ErrorMessage = processErr.Error()
		} else {
			log.Status = domain.PaymentLogProcessed
		}
	default:
		log.Status = domain.PaymentLogIgnored
	}
	_ = h.paymentLogRepo.Update(log)

	c.JSON(http.StatusOK, gin.H{"received": true})
}

// ── Admin payment log endpoints ────────────────────────────────────────────────

func (h *OrderHandler) ListPaymentLogs(c *gin.Context) {
	gateway := c.Query("gateway")
	status := c.Query("status")
	eventType := c.Query("eventType")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	logs, total, err := h.paymentLogRepo.List(gateway, status, eventType, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": logs, "total": total, "page": page, "pageSize": pageSize})
}

func (h *OrderHandler) DownloadReceipt(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}
	order, err := h.orderSvc.GetOrder(id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	pdfBytes, err := pdf.GenerateReceipt(order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate receipt"})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "inline; filename=\""+order.OrderNo+".pdf\"")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func (h *OrderHandler) GetPaymentLog(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	l, err := h.paymentLogRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	// Expose raw payload only in the detail view
	c.JSON(http.StatusOK, gin.H{
		"id":           l.ID,
		"gateway":      l.Gateway,
		"eventType":    l.EventType,
		"gatewayTxnId": l.GatewayTxnID,
		"orderId":      l.OrderID,
		"orderNo":      l.OrderNo,
		"amountCents":  l.AmountCents,
		"currency":     l.Currency,
		"status":       l.Status,
		"errorMessage": l.ErrorMessage,
		"ipAddress":    l.IPAddress,
		"rawPayload":   l.RawPayload,
		"createdAt":    l.CreatedAt,
	})
}
