package handler

import (
	"net/http"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ShippingHandler struct {
	shippingSvc domain.ShippingService
}

func NewShippingHandler(shippingSvc domain.ShippingService) *ShippingHandler {
	return &ShippingHandler{shippingSvc: shippingSvc}
}

// ── Zones ─────────────────────────────────────────────────────────────────────

func (h *ShippingHandler) ListZones(c *gin.Context) {
	zones, err := h.shippingSvc.ListZones()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, zones)
}

func (h *ShippingHandler) CreateZone(c *gin.Context) {
	var req domain.CreateShippingZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	zone, err := h.shippingSvc.CreateZone(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, zone)
}

func (h *ShippingHandler) UpdateZone(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}
	var req domain.UpdateShippingZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	zone, err := h.shippingSvc.UpdateZone(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, zone)
}

func (h *ShippingHandler) DeleteZone(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}
	if err := h.shippingSvc.DeleteZone(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ── Rates ─────────────────────────────────────────────────────────────────────

func (h *ShippingHandler) ListRates(c *gin.Context) {
	var zoneID *uuid.UUID
	if idStr := c.Query("zoneId"); idStr != "" {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zoneId"})
			return
		}
		zoneID = &id
	}
	rates, err := h.shippingSvc.ListRates(zoneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rates)
}

func (h *ShippingHandler) CreateRate(c *gin.Context) {
	var req domain.CreateShippingRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rate, err := h.shippingSvc.CreateRate(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, rate)
}

func (h *ShippingHandler) UpdateRate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rate ID"})
		return
	}
	var req domain.UpdateShippingRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rate, err := h.shippingSvc.UpdateRate(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rate)
}

func (h *ShippingHandler) DeleteRate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rate ID"})
		return
	}
	if err := h.shippingSvc.DeleteRate(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ── Public: Calculate ────────────────────────────────────────────────────────

func (h *ShippingHandler) CalculateRates(c *gin.Context) {
	var req domain.CalculateShippingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rates, err := h.shippingSvc.CalculateRates(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rates)
}
