package handler

import (
	"net/http"
	"strconv"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/delivery/http/middleware"
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WalletHandler struct {
	walletSvc domain.WalletService
}

func NewWalletHandler(walletSvc domain.WalletService) *WalletHandler {
	return &WalletHandler{walletSvc: walletSvc}
}

// GetMyWallet returns the current user's wallet balance
func (h *WalletHandler) GetMyWallet(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	wallet, err := h.walletSvc.GetOrCreateWallet(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve wallet"})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

// GetMyTransactions returns the current user's wallet transactions
func (h *WalletHandler) GetMyTransactions(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	txs, total, err := h.walletSvc.ListTransactions(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":    txs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// ── Redeem Rewards / Withdrawals (Aameba origin/dev flow) ───────────────────

func (h *WalletHandler) GetWithdrawalInfo(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	info, err := h.walletSvc.GetWithdrawalInfo(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

func (h *WalletHandler) QuoteWithdrawal(c *gin.Context) {
	var req domain.WithdrawalQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	q, err := h.walletSvc.QuoteWithdrawal(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, q)
}

func (h *WalletHandler) CreateWithdrawal(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req domain.CreateWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	w, err := h.walletSvc.CreateWithdrawal(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, w)
}

func (h *WalletHandler) ListMyWithdrawals(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	items, total, err := h.walletSvc.ListMyWithdrawals(userID, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "pageSize": pageSize})
}

func (h *WalletHandler) GetMyWithdrawal(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid withdrawal ID"})
		return
	}
	w, err := h.walletSvc.GetWithdrawal(id, userID, false)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, w)
}

func (h *WalletHandler) AdminListWithdrawals(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	items, total, err := h.walletSvc.ListAllWithdrawals(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "pageSize": pageSize})
}

func (h *WalletHandler) AdminApproveWithdrawal(c *gin.Context) {
	adminID, _ := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid withdrawal ID"})
		return
	}
	var req domain.ReviewWithdrawalRequest
	_ = c.ShouldBindJSON(&req)
	w, err := h.walletSvc.ApproveWithdrawal(id, adminID, req.Remark)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, w)
}

func (h *WalletHandler) AdminRejectWithdrawal(c *gin.Context) {
	adminID, _ := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid withdrawal ID"})
		return
	}
	var req domain.ReviewWithdrawalRequest
	_ = c.ShouldBindJSON(&req)
	w, err := h.walletSvc.RejectWithdrawal(id, adminID, req.Remark)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, w)
}

// ── Admin ──────────────────────────────────────────────────────────────────

// ListWallets (Admin)
func (h *WalletHandler) ListWallets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	search := c.Query("search")
	if page < 1 { page = 1 }
	if pageSize < 1 || pageSize > 100 { pageSize = 10 }

	wallets, total, err := h.walletSvc.ListWallets(page, pageSize, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list wallets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":    wallets,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// AdminAdjustBalance manually credits or debits a user's wallet
func (h *WalletHandler) AdminAdjustBalance(c *gin.Context) {
	userIdStr := c.Param("userId")
	userID, err := uuid.Parse(userIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		Amount      float64 `json:"amount" binding:"required"`
		BalanceType string  `json:"balanceType" binding:"required"` // "RM" or "POINTS"
		Description string  `json:"description" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isCredit := req.Amount > 0
	absAmount := req.Amount
	if !isCredit {
		absAmount = -req.Amount
	}

	if req.BalanceType == "RM" {
		if isCredit {
			_, err = h.walletSvc.CreditRM(userID, absAmount, "admin_adjustment", "", req.Description)
		} else {
			_, err = h.walletSvc.DebitRM(userID, absAmount, "admin_adjustment", "", req.Description)
		}
	} else {
		if isCredit {
			_, err = h.walletSvc.CreditPoints(userID, absAmount, "admin_adjustment", "", req.Description)
		} else {
			_, err = h.walletSvc.DebitPoints(userID, absAmount, "admin_adjustment", "", req.Description)
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "wallet adjusted successfully"})
}

// AdminListTransactions lists transactions, optionally filtered by walletId
func (h *WalletHandler) AdminListTransactions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 { page = 1 }
	if pageSize < 1 || pageSize > 100 { pageSize = 10 }

	var walletIDPtr *uuid.UUID
	if wID := c.Query("walletId"); wID != "" {
		parsed, err := uuid.Parse(wID)
		if err == nil {
			walletIDPtr = &parsed
		}
	}

	txs, total, err := h.walletSvc.ListAllTransactions(walletIDPtr, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":    txs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}
