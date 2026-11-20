package handler

import (
	"net/http"
	"strconv"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/delivery/http/middleware"
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InventoryHandler struct {
	inventorySvc domain.InventoryService
}

func NewInventoryHandler(inventorySvc domain.InventoryService) *InventoryHandler {
	return &InventoryHandler{inventorySvc: inventorySvc}
}

func (h *InventoryHandler) AdjustStock(c *gin.Context) {
	adminID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req domain.StockAdjustRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	movement, err := h.inventorySvc.AdjustStock(adminID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, movement)
}

func (h *InventoryHandler) ListMovements(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	variantIDStr := c.Query("variantId")
	movType := domain.MovementType(c.Query("type"))

	var variantID *uuid.UUID
	if variantIDStr != "" {
		id, err := uuid.Parse(variantIDStr)
		if err == nil {
			variantID = &id
		}
	}

	movements, total, err := h.inventorySvc.ListMovements(variantID, movType, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": movements, "total": total, "page": page, "pageSize": pageSize})
}

func (h *InventoryHandler) GetLowStockItems(c *gin.Context) {
	items, err := h.inventorySvc.GetLowStockItems()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *InventoryHandler) GetSummary(c *gin.Context) {
	summary, err := h.inventorySvc.GetSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *InventoryHandler) BulkReceive(c *gin.Context) {
	adminID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req domain.BulkReceiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	movements, err := h.inventorySvc.BulkReceive(adminID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"movements": movements, "count": len(movements)})
}
