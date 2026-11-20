package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB connects to PostgreSQL and auto-migrates all domain models.
func InitDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}

	db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto;")

	err = db.AutoMigrate(
		&domain.User{},
		&domain.Product{},
		&domain.ProductVariant{},
		&domain.ProductReview{},
		&domain.CartItem{},
		&domain.Order{},
		&domain.OrderItem{},
		&domain.InventoryMovement{},
		&domain.Voucher{},
		&domain.UserVoucher{},
		&domain.VoucherRedemption{},
		&domain.AffiliateCommission{},
		&domain.AffiliateSettings{},
		&domain.RankChangeLog{},
		&domain.WhatsAppConfig{},
		&domain.WhatsAppTemplate{},
		&domain.WhatsAppBlast{},
		&domain.WhatsAppBlastRecipient{},
		&domain.ShippingRate{},
		&domain.AddressBook{},
		&domain.PaymentLog{},
		&domain.Wallet{},
		&domain.WalletTransaction{},
		&domain.WithdrawalRequest{},
		&domain.Banner{},
		&domain.Announcement{},
	)
	if err != nil {
		return nil, fmt.Errorf("auto-migrate: %w", err)
	}

	return db, nil
}

// EnsureDatabase creates the target database if it does not exist.
func EnsureDatabase(host, port, user, password, dbName, sslMode string) error {
	adminDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		host, port, user, password, sslMode,
	)
	sqlDB, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	var exists int
	row := sqlDB.QueryRow("SELECT 1 FROM pg_database WHERE datname=$1", dbName)
	if err := row.Scan(&exists); err != nil {
		_, cerr := sqlDB.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, dbName))
		if cerr != nil {
			return fmt.Errorf("create database %q: %w", dbName, cerr)
		}
	}
	return nil
}

// SeedDefaults ensures Ascend Group demo data exists (idempotent).
func SeedDefaults(db *gorm.DB) error {
	systemAdminID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	leaderID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	sarahID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	jasonID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	mayaID := uuid.MustParse("00000000-0000-0000-0000-000000000005")

	hashBytes, _ := bcrypt.GenerateFromPassword([]byte("Password123"), bcrypt.DefaultCost)
	passHash := string(hashBytes)

	// 1. Seed Admin + 4-Tier MLM Demo Hierarchy (S3 -> S2 -> S1 -> M)
	var adminCount int64
	db.Model(&domain.User{}).Where("email = ?", "admin@ascendgroup.my").Count(&adminCount)
	if adminCount == 0 {
		users := []domain.User{
			{
				ID:            systemAdminID,
				Email:         "admin@ascendgroup.my",
				PasswordHash:  passHash,
				FirstName:     "Ascend",
				LastName:      "Admin",
				Phone:         "0120000001",
				Role:          domain.RoleSuperAdmin,
				IsActive:      true,
				EmailVerified: true,
				ReferralCode:  "ASCENDHQ",
				MemberCode:    "AG00001",
				MemberRank:    domain.RankThreeStar,
				HighestRank:   domain.RankThreeStar,
				Country:       "Malaysia",
			},
			{
				ID:             leaderID,
				Email:          "leader@ascendgroup.my",
				PasswordHash:   passHash,
				FirstName:      "Elena",
				LastName:       "Tan",
				Phone:          "0128880001",
				Role:           domain.RoleCustomer,
				IsActive:       true,
				EmailVerified:  true,
				ReferralCode:   "ELENA888",
				MemberCode:     "AG10001",
				MemberRank:     domain.RankThreeStar,
				HighestRank:    domain.RankThreeStar,
				ReferredByCode: "ASCENDHQ",
				ReferredByID:   &systemAdminID,
				BankName:       "Maybank Berhad",
				BankAccountNo:  "514288019283",
				PayeeName:      "ELENA TAN MEI LING",
				PayeeIDNo:      "900412-14-5522",
				AddressLine1:   "Unit 18-02, Sunway Velocity",
				City:           "Kuala Lumpur",
				State:          "WP Kuala Lumpur",
				Postcode:       "55100",
				Country:        "Malaysia",
			},
			{
				ID:             sarahID,
				Email:          "sarah@ascendgroup.my",
				PasswordHash:   passHash,
				FirstName:      "Sarah",
				LastName:       "Lim",
				Phone:          "0128880002",
				Role:           domain.RoleCustomer,
				IsActive:       true,
				EmailVerified:  true,
				ReferralCode:   "SARAH222",
				MemberCode:     "AG10002",
				MemberRank:     domain.RankTwoStar,
				HighestRank:    domain.RankTwoStar,
				ReferredByCode: "ELENA888",
				ReferredByID:   &leaderID,
				BankName:       "CIMB Bank",
				BankAccountNo:  "704512349876",
				PayeeName:      "SARAH LIM JIA YI",
				PayeeIDNo:      "920815-10-4418",
				AddressLine1:   "No 22, Jalan SS2/66",
				City:           "Petaling Jaya",
				State:          "Selangor",
				Postcode:       "47300",
				Country:        "Malaysia",
			},
			{
				ID:             jasonID,
				Email:          "jason@ascendgroup.my",
				PasswordHash:   passHash,
				FirstName:      "Jason",
				LastName:       "Wong",
				Phone:          "0128880003",
				Role:           domain.RoleCustomer,
				IsActive:       true,
				EmailVerified:  true,
				ReferralCode:   "JASON111",
				MemberCode:     "AG10003",
				MemberRank:     domain.RankOneStar,
				HighestRank:    domain.RankOneStar,
				ReferredByCode: "SARAH222",
				ReferredByID:   &sarahID,
				BankName:       "Public Bank",
				BankAccountNo:  "3190887766",
				PayeeName:      "JASON WONG WEI MING",
				PayeeIDNo:      "940320-07-6611",
				AddressLine1:   "12A, Persiaran Gurney",
				City:           "George Town",
				State:          "Penang",
				Postcode:       "10250",
				Country:        "Malaysia",
			},
			{
				ID:             mayaID,
				Email:          "maya@ascendgroup.my",
				PasswordHash:   passHash,
				FirstName:      "Maya",
				LastName:       "Chong",
				Phone:          "0128880004",
				Role:           domain.RoleCustomer,
				IsActive:       true,
				EmailVerified:  true,
				ReferralCode:   "MAYA0001",
				MemberCode:     "AG10004",
				MemberRank:     domain.RankMember,
				HighestRank:    domain.RankMember,
				ReferredByCode: "JASON111",
				ReferredByID:   &jasonID,
				BankName:       "Hong Leong Bank",
				BankAccountNo:  "22400192837",
				PayeeName:      "MAYA CHONG XIN YI",
				PayeeIDNo:      "961108-14-3328",
				AddressLine1:   "88, Jalan Bukit Bintang",
				City:           "Kuala Lumpur",
				State:          "WP Kuala Lumpur",
				Postcode:       "55100",
				Country:        "Malaysia",
			},
		}
		for _, u := range users {
			db.Create(&u)
		}

		// Seed Wallets + Sample Loyalty Activities for demo accounts
		leaderWallet := domain.Wallet{ID: uuid.New(), UserID: leaderID, PointsBalance: 2480.50}
		sarahWallet := domain.Wallet{ID: uuid.New(), UserID: sarahID, PointsBalance: 1120.00}
		jasonWallet := domain.Wallet{ID: uuid.New(), UserID: jasonID, PointsBalance: 540.00}
		mayaWallet := domain.Wallet{ID: uuid.New(), UserID: mayaID, PointsBalance: 285.00}
		for _, w := range []domain.Wallet{leaderWallet, sarahWallet, jasonWallet, mayaWallet} {
			db.Create(&w)
		}
		sampleTxs := []domain.WalletTransaction{
			{WalletID: leaderWallet.ID, UserID: leaderID, Amount: 645.00, BalanceType: "POINTS", Type: "leadership_bonus", RefID: "AG-202611-DEMO01", OrderNo: "AG-202611-DEMO01", Description: "Leadership Bonus (LRB S3 15%) · Order AG-202611-DEMO01 (Maya Chong)"},
			{WalletID: leaderWallet.ID, UserID: leaderID, Amount: 1200.00, BalanceType: "POINTS", Type: "rebate_bonus", RefID: "AG-202611-DEMO02", OrderNo: "AG-202611-DEMO02", Description: "Sponsor Rebate Bonus (L1) · Order AG-202611-DEMO02 (Sarah Lim)"},
			{WalletID: leaderWallet.ID, UserID: leaderID, Amount: 635.50, BalanceType: "POINTS", Type: "rebate_bonus", RefID: "AG-202611-DEMO03", OrderNo: "AG-202611-DEMO03", Description: "Self Rebate Bonus (RB) · Order AG-202611-DEMO03"},
			{WalletID: mayaWallet.ID, UserID: mayaID, Amount: 165.00, BalanceType: "POINTS", Type: "rebate_bonus", RefID: "AG-202611-DEMO01", OrderNo: "AG-202611-DEMO01", Description: "Self Rebate Bonus (RB) · Order AG-202611-DEMO01"},
			{WalletID: mayaWallet.ID, UserID: mayaID, Amount: 120.00, BalanceType: "POINTS", Type: "rebate_bonus", RefID: "AG-202611-DEMO04", OrderNo: "AG-202611-DEMO04", Description: "Sponsor Rebate Bonus (L1) · Order AG-202611-DEMO04 (Chloe Ng)"},
		}
		for _, tx := range sampleTxs {
			db.Create(&tx)
		}

		// Seed Sample Withdrawal Receipt for Elena
		now := time.Now().AddDate(0, 0, -3)
		db.Create(&domain.WithdrawalRequest{
			ReceiptNo:     "WD-202611-8801",
			UserID:        leaderID,
			Amount:        500.00,
			FeeAmount:     0,
			NetAmount:     500.00,
			Currency:      "MYR",
			BankName:      "Maybank Berhad",
			BankAccountNo: "514288019283",
			PayeeName:     "ELENA TAN MEI LING",
			PayeeIDNo:     "900412-14-5522",
			Status:        "approved",
			AdminRemark:   "Payout transferred via DuitNow",
			ReviewedBy:    &systemAdminID,
			ReviewedAt:    &now,
		})
	}

	// 2. Seed Aameba-aligned AffiliateSettings
	var count int64
	db.Model(&domain.AffiliateSettings{}).Count(&count)
	if count == 0 {
		db.Create(&domain.AffiliateSettings{
			CommissionRate:      15.0,
			MLMLevelRates:       "[15,15,15]",
			LeadershipRates:     `{"S1":15,"S2":30,"S3":45}`,
			MLMDepth:            3,
			OneStarMinDownlines: 10,
			OneStarMinSpend:     8000.0,
			TwoStarMinS1Legs:    5,
			ThreeStarMinS2Legs:  5,
			MinPayoutAmt:        50.0,
			WithdrawalFeeRate:   0,
			IsActive:            true,
		})
	}

	// 3. Seed Claimable & Welcome Vouchers
	var voucherCount int64
	db.Model(&domain.Voucher{}).Count(&voucherCount)
	if voucherCount == 0 {
		exp := time.Now().AddDate(0, 6, 0)
		vouchers := []domain.Voucher{
			{
				Code:               "ASCENDWELCOME",
				Name:               "Welcome First Order RM20 Off",
				Description:        "Enjoy RM20 instant savings on your very first Ascend Group wellness order (min. spend RM120).",
				Type:               domain.VoucherTypeCash,
				Value:              20.00,
				MinSpend:           120.00,
				MaxQty:             1000,
				PerUserLimit:       1,
				IsClaimable:        true,
				FirstPurchaseOnly:  true,
				ValidDaysFromClaim: 30,
				Active:             true,
				ExpiresAt:          &exp,
				CreatedBy:          systemAdminID,
			},
			{
				Code:               "DOPAMINE15",
				Name:               "Vibrant Energy 15% Off",
				Description:        "Celebrate vibrant daily wellness with 15% off any bundle or superfood box (capped at RM50).",
				Type:               domain.VoucherTypePercent,
				Value:              15.00,
				MinSpend:           150.00,
				MaxDiscount:        50.00,
				MaxQty:             500,
				PerUserLimit:       3,
				IsClaimable:        true,
				FirstPurchaseOnly:  false,
				ValidDaysFromClaim: 45,
				Active:             true,
				ExpiresAt:          &exp,
				CreatedBy:          systemAdminID,
			},
			{
				Code:               "ASCEND50VIP",
				Name:               "Ascend Circle RM50 Reward Voucher",
				Description:        "Exclusive RM50 off for Ascend Circle members on orders RM350 and above.",
				Type:               domain.VoucherTypeCash,
				Value:              50.00,
				MinSpend:           350.00,
				MaxQty:             300,
				PerUserLimit:       2,
				IsClaimable:        true,
				FirstPurchaseOnly:  false,
				ValidDaysFromClaim: 60,
				Active:             true,
				ExpiresAt:          &exp,
				CreatedBy:          systemAdminID,
			},
		}
		for _, v := range vouchers {
			db.Create(&v)
		}
	}

	// 4. Seed Announcements
	var annCount int64
	db.Model(&domain.Announcement{}).Count(&annCount)
	if annCount == 0 {
		anns := []domain.Announcement{
			{
				Title:    "Welcome to Ascend Group Loyalty & Circle Rewards!",
				Content:  "Earn 15% Rebate Points on your own purchases and 15% across 2 referral levels, plus up to 45% Leadership Bonus as you climb from Member to Three Star!",
				IsActive: true,
			},
			{
				Title:    "New Dopamine Superfood Series Now Live",
				Content:  "Explore our vibrant new Daily Vitality, Glow Collagen, and Gut Harmony blends — every box earns instant BV & Loyalty Points.",
				IsActive: true,
			},
		}
		for _, a := range anns {
			db.Create(&a)
		}
	}

	// 5. Seed Dopamine Ascend Group Products
	var prodCount int64
	db.Model(&domain.Product{}).Count(&prodCount)
	if prodCount == 0 {
		products := []struct {
			Product domain.Product
			Variant domain.ProductVariant
		}{
			{
				Product: domain.Product{
					SKU:         "AG-VIT-SUN",
					Name:        "Ascend Sunburst Citrus Prebiotic Spark",
					NameZH:      "Ascend 阳光柑橘益生元活力饮",
					Description: "Bursting with Sicilian blood orange, yuzu, and 10g natural inulin prebiotic fiber. Ignites gut balance, bright mood, and clean all-day vitality.",
					Category:    "Daily Vitality",
					Brand:       "Ascend Group",
					ImageURL:    "/images/products/ascend-citrus.jpg",
					Tags:        `["Best Seller", "15% Rebate BV", "Zero Added Sugar", "Halal"]`,
					Status:      domain.ProductStatusPublished,
					IsFeatured:  true,
					SortOrder:   1,
					CreatedBy:   systemAdminID,
				},
				Variant: domain.ProductVariant{
					Name:          "Radiance Box (15 Sachets x 20g)",
					SKU:           "AG-VIT-SUN-15S",
					RetailPrice:   128.00,
					DealerPrice:   98.00,
					BV:            100.00,
					PV:            50.00,
					StockOnHand:   500,
					LowStockAlert: 20,
					Weight:        0.35,
					IsActive:      true,
				},
			},
			{
				Product: domain.Product{
					SKU:         "AG-GLW-BER",
					Name:        "Ascend Berry Bloom Botanical Collagen",
					NameZH:      "Ascend 莓果绽放植物胶原饮",
					Description: "Wild cranberry, acai, and marine tri-peptide botanicals crafted for radiant skin bounce, hormonal harmony, and joyful daily glow.",
					Category:    "Beauty & Glow",
					Brand:       "Ascend Group",
					ImageURL:    "/images/products/ascend-berry.jpg",
					Tags:        `["Glow Favorite", "High BV", "Botanical", "Halal"]`,
					Status:      domain.ProductStatusPublished,
					IsFeatured:  true,
					SortOrder:   2,
					CreatedBy:   systemAdminID,
				},
				Variant: domain.ProductVariant{
					Name:          "Bloom Box (20 Sachets x 15g)",
					SKU:           "AG-GLW-BER-20S",
					RetailPrice:   168.00,
					DealerPrice:   128.00,
					BV:            140.00,
					PV:            70.00,
					StockOnHand:   420,
					LowStockAlert: 20,
					Weight:        0.32,
					IsActive:      true,
				},
			},
			{
				Product: domain.Product{
					SKU:         "AG-GRN-AVO",
					Name:        "Ascend Emerald Matcha & Avocado Superfood",
					NameZH:      "Ascend 翡翠抹茶牛油果超级食物饮",
					Description: "Ceremonial Japanese matcha fused with creamy avocado, organic moringa, and vitamin E for crisp mental focus and cellular nourishment.",
					Category:    "Superfood Blends",
					Brand:       "Ascend Group",
					ImageURL:    "/images/products/ascend-matcha.jpg",
					Tags:        `["Superfood", "Clean Energy", "Matcha", "Halal"]`,
					Status:      domain.ProductStatusPublished,
					IsFeatured:  true,
					SortOrder:   3,
					CreatedBy:   systemAdminID,
				},
				Variant: domain.ProductVariant{
					Name:          "Superfood Box (15 Sachets x 20g)",
					SKU:           "AG-GRN-AVO-15S",
					RetailPrice:   148.00,
					DealerPrice:   112.00,
					BV:            120.00,
					PV:            60.00,
					StockOnHand:   380,
					LowStockAlert: 20,
					Weight:        0.35,
					IsActive:      true,
				},
			},
			{
				Product: domain.Product{
					SKU:         "AG-GUT-VIO",
					Name:        "Ascend Violet Sweet Potato Gut Harmony",
					NameZH:      "Ascend 紫薯花青素肠道平衡饮",
					Description: "Japanese purple sweet potato rich in anthocyanins and prebiotic fiber for gentle digestion, immune resilience, and family-friendly nutrition.",
					Category:    "Gut Health",
					Brand:       "Ascend Group",
					ImageURL:    "/images/products/ascend-violet.jpg",
					Tags:        `["Gut Circle", "Anthocyanin", "Family Favorite"]`,
					Status:      domain.ProductStatusPublished,
					IsFeatured:  true,
					SortOrder:   4,
					CreatedBy:   systemAdminID,
				},
				Variant: domain.ProductVariant{
					Name:          "Harmony Box (15 Sachets x 20g)",
					SKU:           "AG-GUT-VIO-15S",
					RetailPrice:   138.00,
					DealerPrice:   105.00,
					BV:            110.00,
					PV:            55.00,
					StockOnHand:   500,
					LowStockAlert: 20,
					Weight:        0.35,
					IsActive:      true,
				},
			},
			{
				Product: domain.Product{
					SKU:         "AG-BND-STR",
					Name:        "Ascend Starter Glow & Vitality Bundle (3 Boxes)",
					NameZH:      "Ascend 活力启航三盒超值套装",
					Description: "Our signature 3-box Ascend Circle starter kit combining Sunburst Citrus, Berry Bloom, and Emerald Matcha — maximum BV value for fast tier growth!",
					Category:    "Value Bundles",
					Brand:       "Ascend Group",
					ImageURL:    "/images/products/ascend-bundle.jpg",
					Tags:        `["Starter Bundle", "380 BV", "Best Value"]`,
					Status:      domain.ProductStatusPublished,
					IsFeatured:  true,
					SortOrder:   5,
					CreatedBy:   systemAdminID,
				},
				Variant: domain.ProductVariant{
					Name:          "3-Box Starter Bundle (Save RM46)",
					SKU:           "AG-BND-STR-3X",
					RetailPrice:   398.00,
					DealerPrice:   310.00,
					BV:            380.00,
					PV:            180.00,
					StockOnHand:   250,
					LowStockAlert: 15,
					Weight:        1.10,
					IsActive:      true,
				},
			},
		}

		for _, item := range products {
			p := item.Product
			if err := db.Create(&p).Error; err != nil {
				continue
			}
			v := item.Variant
			v.ProductID = p.ID
			_ = db.Create(&v).Error
		}
	}

	// Ensure any previously seeded products using legacy image paths are upgraded to new Ascend Group photos
	legacyImageMap := map[string]string{
		"/images/products/mango-passion.jpg":       "/images/products/ascend-citrus.jpg",
		"/images/products/beaufica-cranberry.jpg":  "/images/products/ascend-berry.jpg",
		"/images/products/avocado-moringa.jpg":     "/images/products/ascend-matcha.jpg",
		"/images/products/purple-sweet-potato.jpg": "/images/products/ascend-violet.jpg",
		"/images/products/cocoa-oat.jpg":           "/images/products/ascend-bundle.jpg",
	}
	for oldURL, newURL := range legacyImageMap {
		db.Model(&domain.Product{}).Where("image_url = ?", oldURL).Update("image_url", newURL)
		db.Model(&domain.ProductVariant{}).Where("image_url = ?", oldURL).Update("image_url", newURL)
	}

	return nil
}
