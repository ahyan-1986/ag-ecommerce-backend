package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/config"
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/delivery/http/handler"
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/delivery/http/router"
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/repository"
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/service"
	pkstorage "github.com/ahyan-1986/ag-ecommerce-backend/pkg/storage"
	pkwhatsapp "github.com/ahyan-1986/ag-ecommerce-backend/pkg/whatsapp"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env if present (ignored in production where env vars are injected)
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Database
	if err := repository.EnsureDatabase(
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	); err != nil {
		log.Fatalf("failed to ensure database: %v", err)
	}
	db, err := repository.InitDB(cfg.Database.GetDSN())
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	// Repositories
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)
	voucherRepo := repository.NewVoucherRepository(db)
	affiliateRepo := repository.NewAffiliateRepository(db)
	whatsappRepo := repository.NewWhatsAppRepository(db)
	shippingRepo := repository.NewShippingRepository(db)
	addressBookRepo := repository.NewAddressBookRepository(db)
	paymentLogRepo := repository.NewPaymentLogRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	bannerRepo := repository.NewBannerRepository(db)
	announcementRepo := repository.NewAnnouncementRepository(db)

	// Auto-migrate shipping tables
	if err := repository.MigrateShipping(db); err != nil {
		log.Printf("WARNING: shipping migration failed: %v", err)
	}

	// Storage
	var storageSvc domain.StorageClient
	if cfg.S3.Endpoint != "" && cfg.S3.AccessKey != "" {
		s3Client, err := pkstorage.NewClient(cfg.S3)
		if err != nil {
			log.Printf("WARNING: S3 storage init failed (%v), uploads disabled", err)
			storageSvc = service.NewNoopStorageClient()
		} else {
			storageSvc = s3Client
		}
	} else {
		storageSvc = service.NewNoopStorageClient()
	}

	// Email
	var emailSvc domain.EmailService
	if cfg.Email.Host != "" {
		emailSvc = service.NewEmailService(
			cfg.Email.Host,
			cfg.Email.Port,
			cfg.Email.Username,
			cfg.Email.Password,
			cfg.Email.From,
			cfg.Email.FromName,
		)
	} else {
		emailSvc = service.NewNoopEmailService()
	}

	// AI
	ctx := context.Background()
	var aiSvc domain.AIService
	if cfg.Gemini.APIKey != "" {
		aiSvc, err = service.NewAIService(ctx, cfg.Gemini.APIKey, cfg.Gemini.Model)
		if err != nil {
			log.Printf("WARNING: AI service init failed (%v), using noop", err)
			aiSvc = service.NewNoopAIService()
		}
	} else {
		aiSvc = service.NewNoopAIService()
	}

	// WhatsApp
	waClient := pkwhatsapp.New()

	// Services
	walletSvc := service.NewWalletService(walletRepo, userRepo, affiliateRepo, db)
	affiliateSvc := service.NewAffiliateService(affiliateRepo, userRepo, walletSvc, db)
	authSvc := service.NewAuthService(userRepo, emailSvc, affiliateSvc, cfg)
	userSvc := service.NewUserService(userRepo, storageSvc)
	productSvc := service.NewProductService(productRepo, storageSvc)
	shippingSvc := service.NewShippingService(shippingRepo)
	orderSvc := service.NewOrderService(orderRepo, productRepo, userRepo, voucherRepo, inventoryRepo, affiliateSvc, walletSvc, emailSvc, shippingSvc, cfg)
	inventorySvc := service.NewInventoryService(inventoryRepo, productRepo)
	voucherSvc := service.NewVoucherService(voucherRepo, orderRepo)
	whatsappSvc := service.NewWhatsAppService(whatsappRepo, userRepo, waClient)
	addressBookSvc := service.NewAddressBookService(addressBookRepo)
	bannerSvc := service.NewBannerService(bannerRepo, storageSvc)
	announcementSvc := service.NewAnnouncementService(announcementRepo, storageSvc)

	// Seed defaults
	if err := repository.SeedDefaults(db); err != nil {
		log.Printf("WARNING: seed defaults failed: %v", err)
	}

	// Handlers
	handlers := &router.Handlers{
		Auth:        handler.NewAuthHandler(authSvc),
		User:        handler.NewUserHandler(userSvc),
		Product:     handler.NewProductHandler(productSvc),
		Order:       handler.NewOrderHandler(orderSvc, paymentLogRepo, cfg.Stripe.WebhookSecret),
		Inventory:   handler.NewInventoryHandler(inventorySvc),
		Voucher:     handler.NewVoucherHandler(voucherSvc),
		Affiliate:   handler.NewAffiliateHandler(affiliateSvc, userRepo),
		WhatsApp:    handler.NewWhatsAppHandler(whatsappSvc),
		AI:          handler.NewAIHandler(aiSvc),
		Shipping:    handler.NewShippingHandler(shippingSvc),
		AddressBook: handler.NewAddressBookHandler(addressBookSvc),
		Wallet:       handler.NewWalletHandler(walletSvc),
		Banner:       handler.NewBannerHandler(bannerSvc),
		Announcement: handler.NewAnnouncementHandler(announcementSvc),
	}

	// Router
	r := router.Setup(cfg, handlers)

	// HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("ag-ecommerce-backend listening on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server exited cleanly")
}
