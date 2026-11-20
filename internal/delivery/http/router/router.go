package router

import (
	"net/http"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/config"
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/delivery/http/handler"
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/delivery/http/middleware"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Auth        *handler.AuthHandlerV2
	User        *handler.UserHandler
	Product     *handler.ProductHandler
	Order       *handler.OrderHandler
	Inventory   *handler.InventoryHandler
	Voucher     *handler.VoucherHandler
	Affiliate   *handler.AffiliateHandler
	WhatsApp    *handler.WhatsAppHandler
	AI          *handler.AIHandler
	Shipping    *handler.ShippingHandler
	AddressBook *handler.AddressBookHandler
	Wallet      *handler.WalletHandler
	Banner      *handler.BannerHandler
	Announcement *handler.AnnouncementHandler
}

func Setup(cfg *config.Config, h *Handlers) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	// Health check
	healthHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "ag-ecommerce-backend", "brand": "Ascend Group"})
	}
	r.GET("/health", healthHandler)

	// Stripe webhook (raw body, no auth)
	r.POST("/webhooks/stripe", h.Order.StripeWebhook)

	api := r.Group("/api/v1")
	api.GET("/health", healthHandler)

	// ── Auth (public) ──────────────────────────────────────────────────────
	auth := api.Group("/auth")
	{
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
		auth.POST("/refresh", h.Auth.RefreshToken)
		auth.GET("/verify-email", h.Auth.VerifyEmail)
		auth.POST("/forgot-password", h.Auth.ForgotPassword)
		auth.POST("/reset-password", h.Auth.ResetPassword)
	}

	// ── Public shop (no auth) ──────────────────────────────────────────────
	public := api.Group("/public")
	{
		public.GET("/products", h.Product.ListPublicProducts)
		public.GET("/products/featured", h.Product.ListFeatured)
		public.GET("/products/:id", h.Product.GetPublicProduct)
		public.GET("/products/:id/reviews", h.Product.ListReviews)
		// Voucher validate moved to authenticated /me route
		// Shipping calculator (public)
		public.POST("/shipping/calculate", h.Shipping.CalculateRates)
		public.GET("/banners", h.Banner.ListPublicBanners)
		public.POST("/ai/chat", h.AI.CustomerChat)
		public.GET("/referral/:code", h.Affiliate.VerifyReferralCode)
		public.POST("/guest-checkout", h.Order.GuestCheckout)
	}

	authMW := middleware.AuthMiddleware(cfg)
	adminMW := middleware.RequireAdmin()

	// ── Authenticated customer / dealer routes ─────────────────────────────
	me := api.Group("/me", authMW)
	{
		// Profile
		me.GET("/profile", h.User.GetProfile)
		me.PUT("/profile", h.User.UpdateProfile)
		me.POST("/profile/avatar", h.User.UploadAvatar)
		me.PUT("/password", h.User.ChangePassword)

		// Cart
		me.GET("/cart", h.Order.GetCart)
		me.POST("/cart", h.Order.AddToCart)
		me.PUT("/cart/:itemId", h.Order.UpdateCartItem)
		me.DELETE("/cart/:itemId", h.Order.RemoveCartItem)

		// Orders
		me.POST("/checkout", h.Order.Checkout)
		me.POST("/orders/capture/:orderNo", h.Order.CaptureMyOrderPayment)
		me.GET("/orders/by-no/:orderNo", h.Order.GetMyOrderByNo)
		me.GET("/orders", h.Order.GetMyOrders)
		me.GET("/orders/:id", h.Order.GetMyOrder)
		me.GET("/orders/:id/receipt", h.Order.DownloadReceipt)
		me.POST("/orders/:id/cancel", h.Order.CancelMyOrder)

		// Reviews
		me.POST("/reviews", h.Product.CreateReview)

		// Voucher validation & My Vouchers (Held + Claimable)
		me.POST("/vouchers/validate", h.Voucher.ValidateVoucher)
		me.GET("/vouchers", h.Voucher.GetMyVouchers)
		me.POST("/vouchers/:code/claim", h.Voucher.ClaimVoucher)

		// Announcements
		me.GET("/announcements", h.Announcement.ListActive)

		// Member Dashboard & Affiliate / MLM
		me.GET("/dashboard", h.Affiliate.GetMemberDashboard)
		me.GET("/affiliate/stats", h.Affiliate.GetMyStats)
		me.GET("/affiliate/commissions", h.Affiliate.GetMyCommissions)
		me.GET("/affiliate/network-stats", h.Affiliate.GetMyNetworkStats)
		me.GET("/affiliate/sponsor", h.Affiliate.GetMySponsor)
		me.GET("/affiliate/network-tree", h.Affiliate.GetMyNetworkTree)

		// Wallet & Redeem Rewards (Withdrawals)
		me.GET("/wallet", h.Wallet.GetMyWallet)
		me.GET("/wallet/transactions", h.Wallet.GetMyTransactions)
		me.GET("/withdrawals/info", h.Wallet.GetWithdrawalInfo)
		me.POST("/withdrawals/quote", h.Wallet.QuoteWithdrawal)
		me.POST("/withdrawals", h.Wallet.CreateWithdrawal)
		me.GET("/withdrawals", h.Wallet.ListMyWithdrawals)
		me.GET("/withdrawals/:id", h.Wallet.GetMyWithdrawal)

		// AI chat
		me.POST("/ai/chat", h.AI.CustomerChat)

		// Address book
		me.GET("/addresses", h.AddressBook.List)
		me.POST("/addresses", h.AddressBook.Create)
		me.PUT("/addresses/:id", h.AddressBook.Update)
		me.DELETE("/addresses/:id", h.AddressBook.Delete)
		me.PATCH("/addresses/:id/default", h.AddressBook.SetDefault)
	}

	// ── Admin routes ───────────────────────────────────────────────────────
	admin := api.Group("/admin", authMW, adminMW)
	{
		// Dashboard
		admin.GET("/dashboard", h.Order.GetDashboardStats)

		// Users / CRM
		admin.GET("/users", h.User.ListUsers)
		admin.GET("/users/crm-summary", h.User.GetCRMSummary)
		admin.GET("/users/:id", h.User.GetUser)
		admin.GET("/users/:id/crm-stats", h.User.GetUserCRMStats)
		admin.PUT("/users/:id/active", h.User.SetActive)
		admin.POST("/users/:id/approve-dealer", h.User.ApproveDealer)
		admin.GET("/users/:id/rank-trace", h.Affiliate.GetRankTrace)
		admin.PUT("/users/:id/rank", h.Affiliate.SetMemberRankManual)

		// Products
		admin.GET("/products", h.Product.ListProducts)
		admin.GET("/products/:id", h.Product.GetProduct)
		admin.POST("/products", h.Product.CreateProduct)
		admin.PUT("/products/:id", h.Product.UpdateProduct)
		admin.DELETE("/products/:id", h.Product.DeleteProduct)
		admin.POST("/products/:id/image", h.Product.UploadImage)
		admin.POST("/products/:id/variants", h.Product.CreateVariant)
		admin.PUT("/products/:id/variants/:variantId", h.Product.UpdateVariant)
		admin.DELETE("/products/:id/variants/:variantId", h.Product.DeleteVariant)

		// Orders
		admin.GET("/orders", h.Order.ListOrders)
		admin.GET("/orders/:id", h.Order.GetOrder)
		admin.GET("/orders/:id/receipt", h.Order.DownloadReceipt)
		admin.PUT("/orders/:id/fulfillment", h.Order.UpdateFulfillment)

		// Payment gateway logs
		admin.GET("/payment-logs", h.Order.ListPaymentLogs)
		admin.GET("/payment-logs/:id", h.Order.GetPaymentLog)

		// Inventory
		admin.POST("/inventory/adjust", h.Inventory.AdjustStock)
		admin.POST("/inventory/bulk-receive", h.Inventory.BulkReceive)
		admin.GET("/inventory/movements", h.Inventory.ListMovements)
		admin.GET("/inventory/low-stock", h.Inventory.GetLowStockItems)
		admin.GET("/inventory/summary", h.Inventory.GetSummary)

		// Vouchers
		admin.GET("/vouchers", h.Voucher.ListVouchers)
		admin.GET("/vouchers/redemptions", h.Voucher.ListRedemptions)
		admin.GET("/vouchers/:id", h.Voucher.GetVoucher)
		admin.POST("/vouchers", h.Voucher.CreateVoucher)
		admin.PUT("/vouchers/:id", h.Voucher.UpdateVoucher)
		admin.DELETE("/vouchers/:id", h.Voucher.DeleteVoucher)

		// Affiliate / MLM
		admin.GET("/affiliate/settings", h.Affiliate.GetSettings)
		admin.PUT("/affiliate/settings", h.Affiliate.UpdateSettings)
		admin.GET("/affiliate/commissions", h.Affiliate.ListAllCommissions)
		admin.POST("/affiliate/commissions/:id/approve", h.Affiliate.ApproveCommission)
		admin.POST("/affiliate/commissions/:id/pay", h.Affiliate.MarkCommissionPaid)
		admin.GET("/affiliate/network/:userId", h.Affiliate.GetAdminNetworkTree)

		// WhatsApp
		admin.GET("/whatsapp/config", h.WhatsApp.GetConfig)
		admin.PUT("/whatsapp/config", h.WhatsApp.SaveConfig)
		admin.GET("/whatsapp/templates", h.WhatsApp.ListTemplates)
		admin.POST("/whatsapp/templates", h.WhatsApp.CreateTemplate)
		admin.DELETE("/whatsapp/templates/:id", h.WhatsApp.DeleteTemplate)
		admin.GET("/whatsapp/blasts", h.WhatsApp.ListBlasts)
		admin.GET("/whatsapp/blasts/:id", h.WhatsApp.GetBlast)
		admin.POST("/whatsapp/blasts", h.WhatsApp.CreateBlast)
		admin.POST("/whatsapp/blasts/:id/send", h.WhatsApp.SendBlast)

		// AI analytics
		admin.POST("/ai/analytics", h.AI.AdminAnalytics)

		// Shipping zones & rates
		admin.GET("/shipping/zones", h.Shipping.ListZones)
		admin.POST("/shipping/zones", h.Shipping.CreateZone)
		admin.PUT("/shipping/zones/:id", h.Shipping.UpdateZone)
		admin.DELETE("/shipping/zones/:id", h.Shipping.DeleteZone)
		admin.GET("/shipping/rates", h.Shipping.ListRates)
		admin.POST("/shipping/rates", h.Shipping.CreateRate)
		admin.PUT("/shipping/rates/:id", h.Shipping.UpdateRate)
		admin.DELETE("/shipping/rates/:id", h.Shipping.DeleteRate)

		// Wallets & Withdrawals
		admin.GET("/wallets", h.Wallet.ListWallets)
		admin.GET("/wallets/transactions", h.Wallet.AdminListTransactions)
		admin.POST("/wallets/:userId/adjust", h.Wallet.AdminAdjustBalance)
		admin.GET("/withdrawals", h.Wallet.AdminListWithdrawals)
		admin.POST("/withdrawals/:id/approve", h.Wallet.AdminApproveWithdrawal)
		admin.POST("/withdrawals/:id/reject", h.Wallet.AdminRejectWithdrawal)

		// Banners
		admin.GET("/banners", h.Banner.ListAdminBanners)
		admin.POST("/banners", h.Banner.CreateBanner)
		admin.PUT("/banners/:id", h.Banner.UpdateBanner)
		admin.DELETE("/banners/:id", h.Banner.DeleteBanner)
		admin.POST("/banners/:id/image", h.Banner.UploadBannerImage)

		// Announcements
		admin.GET("/announcements", h.Announcement.ListAdmin)
		admin.POST("/announcements", h.Announcement.Create)
		admin.PUT("/announcements/:id", h.Announcement.Update)
		admin.POST("/announcements/:id/attachment", h.Announcement.UploadAttachment)
		admin.DELETE("/announcements/:id", h.Announcement.Delete)
	}

	return r
}
