package handler

import (
	"io"
	"net/http"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AnnouncementHandler struct {
	svc service.AnnouncementService
}

func NewAnnouncementHandler(svc service.AnnouncementService) *AnnouncementHandler {
	return &AnnouncementHandler{svc}
}

type CreateAnnouncementReq struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	IsActive bool   `json:"isActive"`
	Priority int    `json:"priority"`
}

func (h *AnnouncementHandler) Create(c *gin.Context) {
	var req CreateAnnouncementReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	a, err := h.svc.CreateAnnouncement(req.Title, req.Content, req.IsActive, req.Priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, a)
}

type UpdateAnnouncementReq struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	IsActive bool   `json:"isActive"`
	Priority int    `json:"priority"`
}

func (h *AnnouncementHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdateAnnouncementReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateAnnouncement(id, req.Title, req.Content, req.IsActive, req.Priority); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "announcement updated"})
}

func (h *AnnouncementHandler) UploadAttachment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid announcement id"})
		return
	}
	file, header, err := c.Request.FormFile("attachment")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "attachment file required"})
		return
	}
	defer file.Close()
	if header.Size > 20*1024*1024 { // 20MB limit
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 20MB)"})
		return
	}
	data, _ := io.ReadAll(file)
	url, err := h.svc.UploadAttachment(id, data, header.Header.Get("Content-Type"), header.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"attachmentUrl": url})
}

func (h *AnnouncementHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.svc.DeleteAnnouncement(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "announcement deleted"})
}

func (h *AnnouncementHandler) ListAdmin(c *gin.Context) {
	as, err := h.svc.GetAdminAnnouncements()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, as)
}

func (h *AnnouncementHandler) ListActive(c *gin.Context) {
	as, err := h.svc.GetActiveAnnouncements()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, as)
}
