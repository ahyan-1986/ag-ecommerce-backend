package handler

import (
	"log"
	"net/http"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/delivery/http/middleware"
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	aiSvc domain.AIService
}

func NewAIHandler(aiSvc domain.AIService) *AIHandler {
	return &AIHandler{aiSvc: aiSvc}
}

// CustomerChat handles the customer-facing AI chat endpoint (supports both members and guests).
func (h *AIHandler) CustomerChat(c *gin.Context) {
	var userStr string
	if userID, ok := middleware.GetUserID(c); ok {
		userStr = userID.String()
	} else {
		userStr = "guest"
	}

	var req struct {
		Message   string `json:"message" binding:"required"`
		SessionID string `json:"sessionId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.SessionID == "" {
		req.SessionID = "guest-session"
	}
	reply, err := h.aiSvc.CustomerChat(userStr, req.SessionID, req.Message)
	if err != nil {
		log.Printf("[AI] CustomerChat error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"reply": reply})
}

// AdminAnalytics handles the admin analytics chat (ChatGPT-like window).
func (h *AIHandler) AdminAnalytics(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req struct {
		Prompt    string `json:"prompt" binding:"required"`
		DBContext string `json:"dbContext"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	reply, err := h.aiSvc.AdminAnalytics(userID.String(), req.Prompt, req.DBContext)
	if err != nil {
		log.Printf("[AI] AdminAnalytics error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"reply": reply})
}
