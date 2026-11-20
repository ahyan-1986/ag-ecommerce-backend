package handler

import (
	"net/http"
	"strconv"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/delivery/http/middleware"
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WhatsAppHandler struct {
	whatsappSvc domain.WhatsAppService
}

func NewWhatsAppHandler(whatsappSvc domain.WhatsAppService) *WhatsAppHandler {
	return &WhatsAppHandler{whatsappSvc: whatsappSvc}
}

func (h *WhatsAppHandler) GetConfig(c *gin.Context) {
	cfg, err := h.whatsappSvc.GetConfig()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}
	if cfg != nil {
		cfg.AccessToken = "***"
	}
	c.JSON(http.StatusOK, cfg)
}

func (h *WhatsAppHandler) SaveConfig(c *gin.Context) {
	adminID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req domain.SaveWhatsAppConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cfg, err := h.whatsappSvc.SaveConfig(adminID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cfg.AccessToken = "***"
	c.JSON(http.StatusOK, cfg)
}

func (h *WhatsAppHandler) ListTemplates(c *gin.Context) {
	templates, err := h.whatsappSvc.ListTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, templates)
}

func (h *WhatsAppHandler) CreateTemplate(c *gin.Context) {
	var req domain.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, err := h.whatsappSvc.CreateTemplate(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, t)
}

func (h *WhatsAppHandler) DeleteTemplate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}
	if err := h.whatsappSvc.DeleteTemplate(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "template deleted"})
}

func (h *WhatsAppHandler) ListBlasts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	blasts, total, err := h.whatsappSvc.ListBlasts(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": blasts, "total": total, "page": page, "pageSize": pageSize})
}

func (h *WhatsAppHandler) GetBlast(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid blast ID"})
		return
	}
	blast, err := h.whatsappSvc.GetBlast(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "blast not found"})
		return
	}
	c.JSON(http.StatusOK, blast)
}

func (h *WhatsAppHandler) CreateBlast(c *gin.Context) {
	adminID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req domain.CreateBlastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	blast, err := h.whatsappSvc.CreateBlast(adminID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, blast)
}

func (h *WhatsAppHandler) SendBlast(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid blast ID"})
		return
	}
	if err := h.whatsappSvc.SendBlast(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "blast queued for sending"})
}
