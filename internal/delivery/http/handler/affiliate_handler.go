package handler

import (
	"net/http"
	"strconv"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/delivery/http/middleware"
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AffiliateHandler struct {
	affiliateSvc domain.AffiliateService
	userRepo     domain.UserRepository
}

func NewAffiliateHandler(affiliateSvc domain.AffiliateService, userRepo domain.UserRepository) *AffiliateHandler {
	return &AffiliateHandler{affiliateSvc: affiliateSvc, userRepo: userRepo}
}

// ── Customer endpoints ────────────────────────────────────────────────────────

// GetMyStats returns simple stats (backward compat).
func (h *AffiliateHandler) GetMyStats(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	stats, err := h.affiliateSvc.GetMyStats(userID, user.ReferralCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetMyNetworkStats returns full MLM dashboard stats for the logged-in user.
func (h *AffiliateHandler) GetMyNetworkStats(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	stats, err := h.affiliateSvc.GetMyNetworkStats(userID, user.ReferralCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetMySponsor returns the direct upline sponsor info.
func (h *AffiliateHandler) GetMySponsor(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	info, err := h.affiliateSvc.GetMySponsor(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

// GetMyNetworkTree returns the recursive downline tree for the logged-in user.
func (h *AffiliateHandler) GetMyNetworkTree(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	tree, err := h.affiliateSvc.GetMyNetworkTree(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tree)
}

// GetMemberDashboard returns the Aameba-style member dashboard payload (tier, points, referrals, rank ladder).
func (h *AffiliateHandler) GetMemberDashboard(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	dash, err := h.affiliateSvc.GetMemberDashboard(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dash)
}

// VerifyReferralCode (public) verifies a referral code for /register/:referralCode or checkout.
func (h *AffiliateHandler) VerifyReferralCode(c *gin.Context) {
	code := c.Param("code")
	user, err := h.userRepo.FindByReferralCode(code)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"valid": false, "error": "referral code not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"valid":        true,
		"referralCode": user.ReferralCode,
		"memberCode":   user.MemberCode,
		"sponsorName":  user.FirstName + " " + user.LastName,
		"memberRank":   user.HighestRank,
	})
}

// GetRankTrace (admin) returns rank change audit history and direct downline purchase summary.
func (h *AffiliateHandler) GetRankTrace(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	trace, err := h.affiliateSvc.GetRankTrace(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, trace)
}

// SetMemberRankManual (admin) updates a member's rank with audit log.
func (h *AffiliateHandler) SetMemberRankManual(c *gin.Context) {
	adminID, _ := middleware.GetUserID(c)
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	var req struct {
		Rank domain.MemberRank `json:"rank" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	adminLabel := "admin"
	if adminUser, err := h.userRepo.FindByID(adminID); err == nil && adminUser != nil {
		adminLabel = adminUser.Email
	}
	if err := h.affiliateSvc.SetMemberRankManual(userID, req.Rank, adminLabel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member rank updated"})
}

// GetMyCommissions lists paginated commissions for the logged-in user.
func (h *AffiliateHandler) GetMyCommissions(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	commissions, total, err := h.affiliateSvc.ListMyCommissions(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": commissions, "total": total, "page": page, "pageSize": pageSize})
}

// ── Admin endpoints ───────────────────────────────────────────────────────────

// GetSettings returns affiliate/MLM settings.
func (h *AffiliateHandler) GetSettings(c *gin.Context) {
	settings, err := h.affiliateSvc.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

// UpdateSettings updates commission rates, depth, and activation.
func (h *AffiliateHandler) UpdateSettings(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req domain.UpdateAffiliateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	settings, err := h.affiliateSvc.UpdateSettings(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

// ListAllCommissions lists all commissions (admin).
func (h *AffiliateHandler) ListAllCommissions(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	commissions, total, err := h.affiliateSvc.ListAllCommissions(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": commissions, "total": total, "page": page, "pageSize": pageSize})
}

// ApproveCommission approves a pending commission (admin).
func (h *AffiliateHandler) ApproveCommission(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid commission ID"})
		return
	}
	if err := h.affiliateSvc.ApproveCommission(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "commission approved"})
}

// MarkCommissionPaid marks a commission as paid (admin).
func (h *AffiliateHandler) MarkCommissionPaid(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid commission ID"})
		return
	}
	if err := h.affiliateSvc.MarkCommissionPaid(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "commission marked as paid"})
}

// GetAdminNetworkTree returns the full downline tree for any user (admin).
func (h *AffiliateHandler) GetAdminNetworkTree(c *gin.Context) {
	rootIDStr := c.Param("userId")
	rootID, err := uuid.Parse(rootIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	tree, err := h.affiliateSvc.GetAdminNetworkTree(rootID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tree)
}
