package handler

import (
	"net/http"
	"strconv"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/delivery/http/middleware"
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type VoucherHandler struct {
	voucherSvc domain.VoucherService
}

func NewVoucherHandler(voucherSvc domain.VoucherService) *VoucherHandler {
	return &VoucherHandler{voucherSvc: voucherSvc}
}

func (h *VoucherHandler) ListVouchers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	var active *bool
	if activeStr := c.Query("active"); activeStr != "" {
		b := activeStr == "true"
		active = &b
	}
	vouchers, total, err := h.voucherSvc.ListVouchers(active, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": vouchers, "total": total, "page": page, "pageSize": pageSize})
}

func (h *VoucherHandler) GetVoucher(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid voucher ID"})
		return
	}
	v, err := h.voucherSvc.GetVoucher(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "voucher not found"})
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *VoucherHandler) CreateVoucher(c *gin.Context) {
	adminID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req domain.CreateVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := h.voucherSvc.CreateVoucher(adminID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

func (h *VoucherHandler) UpdateVoucher(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid voucher ID"})
		return
	}
	var req domain.UpdateVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := h.voucherSvc.UpdateVoucher(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *VoucherHandler) DeleteVoucher(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid voucher ID"})
		return
	}
	if err := h.voucherSvc.DeleteVoucher(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "voucher deleted"})
}

func (h *VoucherHandler) ValidateVoucher(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req domain.ValidateVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.voucherSvc.ValidateVoucher(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *VoucherHandler) ListRedemptions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	var voucherID *uuid.UUID
	if idStr := c.Query("voucherId"); idStr != "" {
		id, err := uuid.Parse(idStr)
		if err == nil {
			voucherID = &id
		}
	}
	items, total, err := h.voucherSvc.ListRedemptions(voucherID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "pageSize": pageSize})
}

func (h *VoucherHandler) GetMyVouchers(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	resp, err := h.voucherSvc.GetMyVouchers(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *VoucherHandler) ClaimVoucher(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	code := c.Param("code")
	uv, err := h.voucherSvc.ClaimVoucher(userID, code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, uv)
}
