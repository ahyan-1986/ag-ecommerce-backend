package service

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type affiliateService struct {
	affiliateRepo domain.AffiliateRepository
	userRepo      domain.UserRepository
	walletSvc     domain.WalletService
	db            *gorm.DB
}

func NewAffiliateService(
	affiliateRepo domain.AffiliateRepository,
	userRepo domain.UserRepository,
	walletSvc domain.WalletService,
	db *gorm.DB,
) domain.AffiliateService {
	return &affiliateService{
		affiliateRepo: affiliateRepo,
		userRepo:      userRepo,
		walletSvc:     walletSvc,
		db:            db,
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func parseLevelRates(raw string) []float64 {
	var rates []float64
	if err := json.Unmarshal([]byte(raw), &rates); err != nil || len(rates) == 0 {
		return []float64{15, 15, 15} // Aameba origin/dev default: L0=15%, L1=15%, L2=15%
	}
	return rates
}

func parseLeadershipRates(raw string) map[string]float64 {
	rates := map[string]float64{"S1": 15, "S2": 30, "S3": 45}
	if raw == "" {
		return rates
	}
	var parsed map[string]float64
	if err := json.Unmarshal([]byte(raw), &parsed); err == nil && len(parsed) > 0 {
		return parsed
	}
	return rates
}

func (s *affiliateService) GetSettings() (*domain.AffiliateSettings, error) {
	return s.affiliateRepo.GetSettings()
}

func (s *affiliateService) UpdateSettings(adminID uuid.UUID, req domain.UpdateAffiliateSettingsRequest) (*domain.AffiliateSettings, error) {
	settings, err := s.affiliateRepo.GetSettings()
	if err != nil {
		return nil, err
	}
	settings.UpdatedBy = adminID
	settings.IsActive = req.IsActive
	if req.MinPayoutAmt > 0 {
		settings.MinPayoutAmt = req.MinPayoutAmt
	}
	settings.WithdrawalFeeRate = req.WithdrawalFeeRate

	if len(req.MLMLevelRates) > 0 {
		b, _ := json.Marshal(req.MLMLevelRates)
		settings.MLMLevelRates = string(b)
		settings.CommissionRate = req.MLMLevelRates[0]
	} else if req.CommissionRate > 0 {
		settings.CommissionRate = req.CommissionRate
	}
	if len(req.LeadershipRates) > 0 {
		b, _ := json.Marshal(req.LeadershipRates)
		settings.LeadershipRates = string(b)
	}
	if req.MLMDepth > 0 {
		settings.MLMDepth = req.MLMDepth
	}
	if req.OneStarMinDownlines > 0 {
		settings.OneStarMinDownlines = req.OneStarMinDownlines
	}
	if req.OneStarMinSpend > 0 {
		settings.OneStarMinSpend = req.OneStarMinSpend
	}
	if req.TwoStarMinS1Legs > 0 {
		settings.TwoStarMinS1Legs = req.TwoStarMinS1Legs
	}
	if req.ThreeStarMinS2Legs > 0 {
		settings.ThreeStarMinS2Legs = req.ThreeStarMinS2Legs
	}

	return settings, s.affiliateRepo.UpdateSettings(settings)
}

// ProcessOrderCommission executes the exact Aameba origin/dev bonus calculation:
// 1. Rebate Bonus (RB): L0 (self) = 15%, L1 (sponsor) = 15%, L2 (2nd upline) = 15% of Order BV.
//    Inactive upline links are compressed (skipped to next active upline).
// 2. Leadership Rebate Bonus (LRB): Differential breakaway climbing upline:
//    S1 = 15%, S2 = 30%, S3 = 45%. Pays eligibleRate = rankRate - paidRate.
// 3. Guest Orders: Sponsor chain receives RB (L1, L2) + LRB immediately; guest's own 15% self-rebate
//    is held in order.GuestPointsHeld until registration.
// 4. Rank Promotion: Evaluates One Star (S1) for sponsor and cascades S2/S3 upward.
func (s *affiliateService) ProcessOrderCommission(order *domain.Order) error {
	settings, err := s.affiliateRepo.GetSettings()
	if err != nil || !settings.IsActive {
		return nil
	}

	// Idempotency guard
	if _, err := s.affiliateRepo.FindCommissionByOrder(order.ID); err == nil {
		return nil
	}

	orderBV := order.TotalBV
	if orderBV <= 0 {
		orderBV = order.SubTotal
	}
	if orderBV <= 0 {
		return nil
	}

	rebateRates := parseLevelRates(settings.MLMLevelRates)
	leadershipRates := parseLeadershipRates(settings.LeadershipRates)
	now := time.Now()

	// ── GUEST ORDER PATH (GuestSponsorRebate + GuestSponsorLeadership) ────
	if order.BuyerType == "G" {
		sponsorCode := strings.TrimSpace(order.SponsorCode)
		if sponsorCode == "" {
			return nil
		}
		sponsor, err := s.userRepo.FindByReferralCode(sponsorCode)
		if err != nil || sponsor == nil {
			return nil
		}
		giverName := order.GuestName
		if giverName == "" {
			giverName = order.GuestEmail
		}
		if giverName == "" {
			giverName = "Guest"
		}

		// Build upline chain starting from sponsor itself
		chain := []domain.User{*sponsor}
		above, _ := s.affiliateRepo.GetUplineChain(sponsor.ID, 15)
		chain = append(chain, above...)

		// Guest Sponsor Rebate (L1, L2)
		level := 1
		for _, upline := range chain {
			if level >= len(rebateRates) {
				break
			}
			if !upline.IsActive {
				continue // active compression
			}
			rate := rebateRates[level]
			if rate > 0 {
				amt := round2(orderBV * (rate / 100.0))
				if amt > 0 {
					c := &domain.AffiliateCommission{
						ReferrerID:     upline.ID,
						OrderID:        order.ID,
						OrderNo:        order.OrderNo,
						BonusType:      domain.BonusTypeRB,
						GiverName:      giverName,
						EarnerRank:     upline.HighestRank,
						SaleAmount:     order.TotalAmount,
						OrderBV:        orderBV,
						CommissionRate: rate,
						CommissionAmt:  amt,
						Currency:       "MYR",
						Status:         "paid",
						Level:          level,
						Remark:         "#G" + order.OrderNo,
						PaidAt:         &now,
					}
					if err := s.affiliateRepo.CreateCommission(c); err == nil {
						desc := fmt.Sprintf("Guest Sponsor Rebate (L%d) · Order %s (%s)", level, order.OrderNo, giverName)
						_, _ = s.walletSvc.CreditPoints(upline.ID, amt, "rebate_bonus", order.OrderNo, desc)
					}
				}
			}
			level++
		}

		// Guest Sponsor Leadership (Differential LRB)
		s.payLeadershipChain(chain, order, nil, giverName, orderBV, leadershipRates, "#G"+order.OrderNo)
		return nil
	}

	// ── MEMBER ORDER PATH ─────────────────────────────────────────────────
	buyer, err := s.userRepo.FindByID(order.UserID)
	if err != nil || buyer == nil {
		return nil
	}
	giverName := strings.TrimSpace(buyer.FirstName + " " + buyer.LastName)
	if giverName == "" {
		giverName = buyer.Email
	}

	// 1) Self Rebate Bonus (Level 0)
	if len(rebateRates) > 0 && rebateRates[0] > 0 && buyer.IsActive {
		rate := rebateRates[0]
		amt := round2(orderBV * (rate / 100.0))
		if amt > 0 {
			c := &domain.AffiliateCommission{
				ReferrerID:     buyer.ID,
				RefereeID:      &buyer.ID,
				OrderID:        order.ID,
				OrderNo:        order.OrderNo,
				BonusType:      domain.BonusTypeRB,
				GiverName:      giverName,
				EarnerRank:     buyer.HighestRank,
				SaleAmount:     order.TotalAmount,
				OrderBV:        orderBV,
				CommissionRate: rate,
				CommissionAmt:  amt,
				Currency:       "MYR",
				Status:         "paid",
				Level:          0,
				Remark:         "Self Rebate",
				PaidAt:         &now,
			}
			if err := s.affiliateRepo.CreateCommission(c); err == nil {
				desc := fmt.Sprintf("Self Rebate Bonus (RB) · Order %s", order.OrderNo)
				_, _ = s.walletSvc.CreditPoints(buyer.ID, amt, "rebate_bonus", order.OrderNo, desc)
			}
		}
	}

	// Fetch upline chain (up to 20 levels for differential leadership)
	uplineChain, _ := s.affiliateRepo.GetUplineChain(buyer.ID, 20)

	// 2) Upline Rebate Bonus (Level 1 = Direct Sponsor, Level 2 = 2nd Upline) with active compression
	level := 1
	for _, upline := range uplineChain {
		if level >= len(rebateRates) {
			break
		}
		if !upline.IsActive {
			continue // compress inactive upline
		}
		rate := rebateRates[level]
		if rate > 0 {
			amt := round2(orderBV * (rate / 100.0))
			if amt > 0 {
				c := &domain.AffiliateCommission{
					ReferrerID:     upline.ID,
					RefereeID:      &buyer.ID,
					OrderID:        order.ID,
					OrderNo:        order.OrderNo,
					BonusType:      domain.BonusTypeRB,
					GiverName:      giverName,
					EarnerRank:     upline.HighestRank,
					SaleAmount:     order.TotalAmount,
					OrderBV:        orderBV,
					CommissionRate: rate,
					CommissionAmt:  amt,
					Currency:       "MYR",
					Status:         "paid",
					Level:          level,
					Remark:         fmt.Sprintf("L%d Upline Rebate", level),
					PaidAt:         &now,
				}
				if err := s.affiliateRepo.CreateCommission(c); err == nil {
					desc := fmt.Sprintf("Sponsor Rebate Bonus (L%d) · Order %s (%s)", level, order.OrderNo, giverName)
					_, _ = s.walletSvc.CreditPoints(upline.ID, amt, "rebate_bonus", order.OrderNo, desc)
				}
			}
		}
		level++
	}

	// 3) Differential Leadership Rebate Bonus (LRB)
	s.payLeadershipChain(uplineChain, order, &buyer.ID, giverName, orderBV, leadershipRates, "")

	// 4) Evaluate & Promote Ranks (One Star -> Two Star -> Three Star)
	if buyer.ReferredByID != nil {
		_ = s.EvaluateAndPromoteRanks(*buyer.ReferredByID, "PAYMENT_PROMOTION")
	}

	return nil
}

func (s *affiliateService) payLeadershipChain(
	uplineChain []domain.User,
	order *domain.Order,
	refereeID *uuid.UUID,
	giverName string,
	orderBV float64,
	leadershipRates map[string]float64,
	remark string,
) {
	paidRate := 0.0
	pass := 1
	now := time.Now()

	for _, upline := range uplineChain {
		if paidRate >= 45.0 {
			break
		}
		rank := upline.HighestRank
		if rank == "" {
			rank = upline.MemberRank
		}
		rankRate := leadershipRates[string(rank)]
		eligibleRate := 0.0
		if rankRate > paidRate {
			eligibleRate = rankRate - paidRate
		}
		if eligibleRate > 0 && upline.IsActive {
			amt := round2(orderBV * (eligibleRate / 100.0))
			if amt > 0 {
				c := &domain.AffiliateCommission{
					ReferrerID:     upline.ID,
					RefereeID:      refereeID,
					OrderID:        order.ID,
					OrderNo:        order.OrderNo,
					BonusType:      domain.BonusTypeLRB,
					GiverName:      giverName,
					EarnerRank:     rank,
					SaleAmount:     order.TotalAmount,
					OrderBV:        orderBV,
					CommissionRate: eligibleRate,
					CommissionAmt:  amt,
					Currency:       "MYR",
					Status:         "paid",
					Level:          pass,
					Remark:         remark,
					PaidAt:         &now,
				}
				if err := s.affiliateRepo.CreateCommission(c); err == nil {
					desc := fmt.Sprintf("Leadership Bonus (LRB %s %.0f%%) · Order %s (%s)", rank, eligibleRate, order.OrderNo, giverName)
					_, _ = s.walletSvc.CreditPoints(upline.ID, amt, "leadership_bonus", order.OrderNo, desc)
				}
			}
			paidRate = rankRate
			pass++
		}
	}
}

// ReleaseGuestHeldPoints mirrors Aameba GuestRebate: when a guest registers with the same email,
// any held self-rebate points from their previous guest orders are credited to their Loyalty Points wallet.
func (s *affiliateService) ReleaseGuestHeldPoints(user *domain.User) error {
	if s.db == nil || user == nil || user.Email == "" {
		return nil
	}
	var guestOrders []domain.Order
	if err := s.db.Where("buyer_type = 'G' AND LOWER(guest_email) = LOWER(?) AND guest_points_held > 0 AND guest_points_paid = false", user.Email).
		Find(&guestOrders).Error; err != nil {
		return err
	}
	now := time.Now()
	for _, o := range guestOrders {
		amt := round2(o.GuestPointsHeld)
		if amt <= 0 {
			continue
		}
		c := &domain.AffiliateCommission{
			ReferrerID:     user.ID,
			RefereeID:      &user.ID,
			OrderID:        o.ID,
			OrderNo:        o.OrderNo,
			BonusType:      domain.BonusTypeRB,
			GiverName:      strings.TrimSpace(user.FirstName + " " + user.LastName),
			EarnerRank:     user.HighestRank,
			SaleAmount:     o.TotalAmount,
			OrderBV:        o.TotalBV,
			CommissionRate: 15,
			CommissionAmt:  amt,
			Currency:       "MYR",
			Status:         "paid",
			Level:          0,
			Remark:         "Guest Held Rebate Released",
			PaidAt:         &now,
		}
		_ = s.affiliateRepo.CreateCommission(c)
		desc := fmt.Sprintf("Guest Order Self Rebate Released · Order %s", o.OrderNo)
		_, _ = s.walletSvc.CreditPoints(user.ID, amt, "guest_rebate", o.OrderNo, desc)
		s.db.Model(&domain.Order{}).Where("id = ?", o.ID).Updates(map[string]interface{}{
			"guest_points_paid": true,
			"user_id":           user.ID,
		})
	}
	return nil
}

// EvaluateAndPromoteRanks checks One Star (S1), Two Star (S2), and Three Star (S3) criteria
// for userID and cascades upward through their sponsor chain.
func (s *affiliateService) EvaluateAndPromoteRanks(userID uuid.UUID, source string) error {
	settings, err := s.affiliateRepo.GetSettings()
	if err != nil {
		return err
	}
	minDownlines := settings.OneStarMinDownlines
	if minDownlines <= 0 {
		minDownlines = 10
	}
	minSpend := settings.OneStarMinSpend
	if minSpend <= 0 {
		minSpend = 8000
	}
	minS1Legs := settings.TwoStarMinS1Legs
	if minS1Legs <= 0 {
		minS1Legs = 5
	}
	minS2Legs := settings.ThreeStarMinS2Legs
	if minS2Legs <= 0 {
		minS2Legs = 5
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return nil
	}

	current := user.HighestRank
	if current == "" {
		current = domain.RankMember
	}
	next := current

	// Check One Star (S1): >= 10 direct spending downlines AND >= RM 8,000 combined spend
	if next.Ordinal() < domain.RankOneStar.Ordinal() {
		purchases, err := s.affiliateRepo.GetDirectDownlinePurchases(user.ID)
		if err == nil {
			qualified := 0
			combined := 0.0
			for _, p := range purchases {
				if p.TotalSpend > 0 {
					qualified++
					combined += p.TotalSpend
				}
			}
			if qualified >= minDownlines && combined >= minSpend {
				next = domain.RankOneStar
			}
		}
	}

	// Check Two Star (S2): >= 5 direct legs with S1+
	if next.Ordinal() >= domain.RankOneStar.Ordinal() && next.Ordinal() < domain.RankTwoStar.Ordinal() {
		s1Legs, err := s.affiliateRepo.CountLegsWithMinRank(user.ID, domain.RankOneStar)
		if err == nil && s1Legs >= minS1Legs {
			next = domain.RankTwoStar
		}
	}

	// Check Three Star (S3): >= 5 direct legs with S2+
	if next.Ordinal() >= domain.RankTwoStar.Ordinal() && next.Ordinal() < domain.RankThreeStar.Ordinal() {
		s2Legs, err := s.affiliateRepo.CountLegsWithMinRank(user.ID, domain.RankTwoStar)
		if err == nil && s2Legs >= minS2Legs {
			next = domain.RankThreeStar
		}
	}

	if next.Ordinal() > current.Ordinal() {
		user.MemberRank = next
		user.HighestRank = next
		if err := s.userRepo.Update(user); err == nil {
			_ = s.affiliateRepo.CreateRankChangeLog(&domain.RankChangeLog{
				UserID:   user.ID,
				FromRank: current,
				ToRank:   next,
				Source:   source,
				DoneBy:   "SYSTEM",
			})
			// Cascade upward: sponsor may now qualify for S2 or S3!
			if user.ReferredByID != nil {
				_ = s.EvaluateAndPromoteRanks(*user.ReferredByID, source)
			}
		}
	}
	return nil
}

func (s *affiliateService) SetMemberRankManual(userID uuid.UUID, newRank domain.MemberRank, adminEmail string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	oldRank := user.MemberRank
	if oldRank == "" {
		oldRank = domain.RankMember
	}
	user.MemberRank = newRank
	if newRank.Ordinal() > user.HighestRank.Ordinal() {
		user.HighestRank = newRank
	}
	if err := s.userRepo.Update(user); err != nil {
		return err
	}
	_ = s.affiliateRepo.CreateRankChangeLog(&domain.RankChangeLog{
		UserID:   user.ID,
		FromRank: oldRank,
		ToRank:   newRank,
		Source:   "ADMIN_MANUAL",
		DoneBy:   adminEmail,
	})
	return nil
}

func (s *affiliateService) GetRankProgress(userID uuid.UUID) (*domain.RankProgressResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	settings, _ := s.affiliateRepo.GetSettings()
	minDownlines := 10
	minSpend := 8000.0
	minS1Legs := 5
	minS2Legs := 5
	if settings != nil {
		if settings.OneStarMinDownlines > 0 {
			minDownlines = settings.OneStarMinDownlines
		}
		if settings.OneStarMinSpend > 0 {
			minSpend = settings.OneStarMinSpend
		}
		if settings.TwoStarMinS1Legs > 0 {
			minS1Legs = settings.TwoStarMinS1Legs
		}
		if settings.ThreeStarMinS2Legs > 0 {
			minS2Legs = settings.ThreeStarMinS2Legs
		}
	}

	cur := user.HighestRank
	if cur == "" {
		cur = domain.RankMember
	}
	ladder := []domain.MemberRank{domain.RankMember, domain.RankOneStar, domain.RankTwoStar, domain.RankThreeStar}
	rungs := make([]domain.RankRung, 0, 4)
	for _, code := range ladder {
		rungs = append(rungs, domain.RankRung{
			Code:    code,
			Label:   code.Label(),
			Earned:  code.Ordinal() <= cur.Ordinal(),
			Current: code == cur,
		})
	}

	resp := &domain.RankProgressResponse{
		CurrentRank:      cur,
		CurrentRankLabel: cur.Label(),
		Rungs:            rungs,
		Requirements:     []domain.RankRequirement{},
	}

	pct := func(c, t float64) int {
		if t <= 0 {
			return 100
		}
		p := int(math.Round((c / t) * 100))
		if p > 100 {
			p = 100
		}
		if p < 0 {
			p = 0
		}
		return p
	}

	switch cur {
	case domain.RankMember:
		nr := domain.RankOneStar
		resp.NextRank = &nr
		resp.NextRankLabel = nr.Label()
		purchases, _ := s.affiliateRepo.GetDirectDownlinePurchases(userID)
		qualified := 0
		combined := 0.0
		for _, p := range purchases {
			if p.TotalSpend > 0 {
				qualified++
				combined += p.TotalSpend
			}
		}
		resp.Requirements = []domain.RankRequirement{
			{
				Metric:  "SPENDING_DOWNLINES",
				Label:   "Direct Downlines with Purchases",
				Kind:    "count",
				Current: float64(qualified),
				Target:  float64(minDownlines),
				Percent: pct(float64(qualified), float64(minDownlines)),
			},
			{
				Metric:  "COMBINED_SPEND",
				Label:   "Combined Direct Downline Spend",
				Kind:    "money",
				Current: round2(combined),
				Target:  minSpend,
				Percent: pct(combined, minSpend),
			},
		}
	case domain.RankOneStar:
		nr := domain.RankTwoStar
		resp.NextRank = &nr
		resp.NextRankLabel = nr.Label()
		legs, _ := s.affiliateRepo.CountLegsWithMinRank(userID, domain.RankOneStar)
		resp.Requirements = []domain.RankRequirement{
			{
				Metric:  "ONE_STAR_LEGS",
				Label:   "Direct Legs Reaching One Star",
				Kind:    "count",
				Current: float64(legs),
				Target:  float64(minS1Legs),
				Percent: pct(float64(legs), float64(minS1Legs)),
			},
		}
	case domain.RankTwoStar:
		nr := domain.RankThreeStar
		resp.NextRank = &nr
		resp.NextRankLabel = nr.Label()
		legs, _ := s.affiliateRepo.CountLegsWithMinRank(userID, domain.RankTwoStar)
		resp.Requirements = []domain.RankRequirement{
			{
				Metric:  "TWO_STAR_LEGS",
				Label:   "Direct Legs Reaching Two Star",
				Kind:    "count",
				Current: float64(legs),
				Target:  float64(minS2Legs),
				Percent: pct(float64(legs), float64(minS2Legs)),
			},
		}
	}

	return resp, nil
}

func (s *affiliateService) GetMemberDashboard(userID uuid.UUID) (*domain.MemberDashboardResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	directCount, _ := s.affiliateRepo.CountDownline(userID, 1)
	totalDownline, _ := s.affiliateRepo.CountDownline(userID, 0)
	wallet, _ := s.walletSvc.GetOrCreateWallet(userID)
	availPts := 0.0
	if wallet != nil {
		availPts = wallet.PointsBalance
	}
	lifetimePts, _ := s.walletSvc.SumLifetimePointsEarned(userID)
	rankProg, err := s.GetRankProgress(userID)
	if err != nil {
		return nil, err
	}
	rank := user.HighestRank
	if rank == "" {
		rank = domain.RankMember
	}
	return &domain.MemberDashboardResponse{
		MemberCode:      user.MemberCode,
		ReferralCode:    user.ReferralCode,
		MemberRank:      rank,
		MemberRankLabel: rank.Label(),
		TotalReferrals:  directCount,
		TotalDownline:   totalDownline,
		AvailablePoints: round2(availPts),
		LifetimePoints:  round2(lifetimePts),
		RankProgress:    *rankProg,
	}, nil
}

func (s *affiliateService) GetRankTrace(userID uuid.UUID) (*domain.MemberRankTraceResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	settings, _ := s.affiliateRepo.GetSettings()
	minD, minS := 10, 8000.0
	if settings != nil {
		if settings.OneStarMinDownlines > 0 {
			minD = settings.OneStarMinDownlines
		}
		if settings.OneStarMinSpend > 0 {
			minS = settings.OneStarMinSpend
		}
	}
	logs, _ := s.affiliateRepo.ListRankChangeLogs(userID)
	purchases, _ := s.affiliateRepo.GetDirectDownlinePurchases(userID)
	qualified := 0
	combined := 0.0
	for _, p := range purchases {
		if p.TotalSpend > 0 {
			qualified++
			combined += p.TotalSpend
		}
	}
	rank := user.HighestRank
	if rank == "" {
		rank = domain.RankMember
	}
	return &domain.MemberRankTraceResponse{
		UserID:           user.ID,
		MemberCode:       user.MemberCode,
		CurrentRank:      rank,
		RankChanges:      logs,
		DirectDownlines:  purchases,
		QualifiedCount:   qualified,
		CombinedSpend:    round2(combined),
		MinDownlines:     minD,
		MinCombinedSpend: minS,
	}, nil
}

func (s *affiliateService) GetMyStats(userID uuid.UUID, referralCode string) (*domain.AffiliateStats, error) {
	totalEarned, err := s.affiliateRepo.SumCommissions(userID, "")
	if err != nil {
		return nil, err
	}
	pendingEarned, _ := s.affiliateRepo.SumCommissions(userID, "pending")
	directCount, _ := s.affiliateRepo.CountDownline(userID, 1)

	return &domain.AffiliateStats{
		TotalEarned:    totalEarned,
		PendingEarned:  pendingEarned,
		TotalReferrals: directCount,
		MyReferralCode: referralCode,
	}, nil
}

func (s *affiliateService) GetMyNetworkStats(userID uuid.UUID, referralCode string) (*domain.MLMNetworkStats, error) {
	settings, _ := s.affiliateRepo.GetSettings()
	levelRates := []float64{15, 15, 15}
	leadershipRates := map[string]float64{"S1": 15, "S2": 30, "S3": 45}
	maxDepth := 3
	if settings != nil {
		levelRates = parseLevelRates(settings.MLMLevelRates)
		leadershipRates = parseLeadershipRates(settings.LeadershipRates)
		maxDepth = settings.MLMDepth
	}

	nodes, err := s.affiliateRepo.GetDownlineFlat(userID, maxDepth)
	if err != nil {
		return nil, err
	}

	totalDownline := int64(len(nodes))
	directReferrals := int64(0)
	downlineByLevel := make(map[int]int64)
	for _, n := range nodes {
		downlineByLevel[n.Level]++
		if n.Level == 1 {
			directReferrals++
		}
	}

	totalEarned, _ := s.affiliateRepo.SumCommissions(userID, "")
	pendingEarned, _ := s.affiliateRepo.SumCommissions(userID, "pending")
	earnedByLevel, _ := s.affiliateRepo.SumCommissionsByLevel(userID)
	conversions, _ := s.affiliateRepo.CountDirectDownlineConversions(userID)

	return &domain.MLMNetworkStats{
		TotalDownline:    totalDownline,
		DirectReferrals:  directReferrals,
		TotalEarned:      round2(totalEarned),
		PendingEarned:    round2(pendingEarned),
		EarnedByLevel:    earnedByLevel,
		DownlineByLevel:  downlineByLevel,
		MyReferralCode:   referralCode,
		LevelRates:       levelRates,
		LeadershipRates:  leadershipRates,
		TotalConversions: conversions,
	}, nil
}

func (s *affiliateService) GetMySponsor(userID uuid.UUID) (*domain.SponsorInfo, error) {
	upline, err := s.affiliateRepo.GetUplineChain(userID, 1)
	if err != nil || len(upline) == 0 {
		return &domain.SponsorInfo{HasSponsor: false}, nil
	}
	sponsor := upline[0]
	rank := sponsor.HighestRank
	if rank == "" {
		rank = domain.RankMember
	}
	return &domain.SponsorInfo{
		HasSponsor:   true,
		SponsorID:    sponsor.ID,
		MemberCode:   sponsor.MemberCode,
		FirstName:    sponsor.FirstName,
		LastName:     sponsor.LastName,
		Email:        sponsor.Email,
		ReferralCode: sponsor.ReferralCode,
		MemberRank:   rank,
		JoinedAt:     sponsor.CreatedAt,
	}, nil
}

func (s *affiliateService) GetMyNetworkTree(userID uuid.UUID) (*domain.MLMTreeNode, error) {
	return s.buildTree(userID, 5)
}

func (s *affiliateService) GetAdminNetworkTree(rootUserID uuid.UUID) (*domain.MLMTreeNode, error) {
	return s.buildTree(rootUserID, 10)
}

func (s *affiliateService) buildTree(rootID uuid.UUID, maxDepth int) (*domain.MLMTreeNode, error) {
	root, err := s.userRepo.FindByID(rootID)
	if err != nil {
		return nil, err
	}

	nodes, err := s.affiliateRepo.GetDownlineFlat(rootID, maxDepth)
	if err != nil {
		return nil, err
	}

	rootRank := root.HighestRank
	if rootRank == "" {
		rootRank = domain.RankMember
	}

	rootNode := &domain.MLMTreeNode{
		UserID:       root.ID,
		MemberCode:   root.MemberCode,
		FirstName:    root.FirstName,
		LastName:     root.LastName,
		Email:        root.Email,
		ReferralCode: root.ReferralCode,
		MemberRank:   rootRank,
		Level:        0,
		JoinedAt:     root.CreatedAt,
		Children:     []*domain.MLMTreeNode{},
	}

	// Enrich direct downlines with order spend
	purchases, _ := s.affiliateRepo.GetDirectDownlinePurchases(rootID)
	spendMap := make(map[uuid.UUID]domain.DirectDownlinePurchase, len(purchases))
	for _, p := range purchases {
		spendMap[p.UserID] = p
	}

	nodeMap := make(map[uuid.UUID]*domain.MLMTreeNode, len(nodes)+1)
	nodeMap[rootID] = rootNode

	for i := range nodes {
		n := &nodes[i]
		p := spendMap[n.UserID]
		treeNode := &domain.MLMTreeNode{
			UserID:       n.UserID,
			MemberCode:   n.MemberCode,
			FirstName:    n.FirstName,
			LastName:     n.LastName,
			Email:        n.Email,
			ReferralCode: n.ReferralCode,
			MemberRank:   n.MemberRank,
			Level:        n.Level,
			JoinedAt:     n.JoinedAt,
			TotalOrders:  p.OrderCount,
			TotalSpend:   p.TotalSpend,
			Children:     []*domain.MLMTreeNode{},
		}
		nodeMap[n.UserID] = treeNode
	}

	for i := range nodes {
		n := &nodes[i]
		child := nodeMap[n.UserID]
		if n.ParentID != nil {
			if parent, ok := nodeMap[*n.ParentID]; ok {
				parent.Children = append(parent.Children, child)
			}
		}
	}

	return rootNode, nil
}

func (s *affiliateService) ListMyCommissions(userID uuid.UUID, page, pageSize int) ([]domain.AffiliateCommission, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	return s.affiliateRepo.ListCommissions(&userID, "", page, pageSize)
}

func (s *affiliateService) ListAllCommissions(status string, page, pageSize int) ([]domain.AffiliateCommission, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.affiliateRepo.ListCommissions(nil, status, page, pageSize)
}

func (s *affiliateService) ApproveCommission(id uuid.UUID) error {
	return s.affiliateRepo.ApproveCommission(id)
}

func (s *affiliateService) MarkCommissionPaid(id uuid.UUID) error {
	comm, err := s.affiliateRepo.GetCommissionByID(id)
	if err != nil {
		return err
	}
	if comm.Status == "paid" {
		return nil
	}
	description := fmt.Sprintf("Bonus (%s L%d) · Order %s", comm.BonusType, comm.Level, comm.OrderNo)
	_, err = s.walletSvc.CreditPoints(comm.ReferrerID, comm.CommissionAmt, "rebate_bonus", comm.OrderNo, description)
	if err != nil {
		return fmt.Errorf("failed to credit wallet points: %w", err)
	}
	return s.affiliateRepo.MarkPaid(id)
}
