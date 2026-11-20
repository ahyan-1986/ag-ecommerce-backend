package handler

import (
	"io"
	"net/http"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BannerHandler struct {
	bannerSvc service.BannerService
}

func NewBannerHandler(bannerSvc service.BannerService) *BannerHandler {
	return &BannerHandler{bannerSvc: bannerSvc}
}

func (h *BannerHandler) ListPublicBanners(c *gin.Context) {
	banners, err := h.bannerSvc.ListBanners(true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch banners"})
		return
	}
	c.JSON(http.StatusOK, banners)
}

func (h *BannerHandler) ListAdminBanners(c *gin.Context) {
	banners, err := h.bannerSvc.ListBanners(false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch banners"})
		return
	}
	c.JSON(http.StatusOK, banners)
}

func (h *BannerHandler) CreateBanner(c *gin.Context) {
	var req domain.Banner
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.bannerSvc.CreateBanner(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create banner"})
		return
	}
	c.JSON(http.StatusCreated, req)
}

func (h *BannerHandler) UpdateBanner(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid banner id"})
		return
	}
	var req domain.Banner
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.bannerSvc.UpdateBanner(id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update banner"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated successfully"})
}

func (h *BannerHandler) DeleteBanner(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid banner id"})
		return
	}
	if err := h.bannerSvc.DeleteBanner(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete banner"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted successfully"})
}

func (h *BannerHandler) UploadBannerImage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid banner id"})
		return
	}
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file required"})
		return
	}
	defer file.Close()
	if header.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 10MB)"})
		return
	}
	data, _ := io.ReadAll(file)
	url, err := h.bannerSvc.UploadBannerImage(id, data, header.Header.Get("Content-Type"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"imageUrl": url})
}
