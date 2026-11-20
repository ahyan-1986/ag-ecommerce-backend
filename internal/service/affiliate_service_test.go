package service

import (
	"errors"
	"testing"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
)

type mockAffiliateRepo struct {
	settings    *domain.AffiliateSettings
	commissions []*domain.AffiliateCommission
	upline      []domain.User
	rankLogs    []*domain.RankChangeLog
}

func (m *mockAffiliateRepo) GetSettings() (*domain.AffiliateSettings, error) { return m.settings, nil }
func (m *mockAffiliateRepo) UpdateSettings(s *domain.AffiliateSettings) error { m.settings = s; return nil }
func (m *mockAffiliateRepo) CreateCommission(c *domain.AffiliateCommission) error {
	m.commissions = append(m.commissions, c)
	return nil
}
func (m *mockAffiliateRepo) FindCommissionByOrder(orderID uuid.UUID) (*domain.AffiliateCommission, error) {
	return nil, errors.New("not found")
}
func (m *mockAffiliateRepo) GetCommissionByID(id uuid.UUID) (*domain.AffiliateCommission, error) { return nil, nil }
func (m *mockAffiliateRepo) ListCommissions(referrerID *uuid.UUID, status string, page, pageSize int) ([]domain.AffiliateCommission, int64, error) {
	return nil, 0, nil
}
func (m *mockAffiliateRepo) SumCommissions(referrerID uuid.UUID, status string) (float64, error) { return 0, nil }
func (m *mockAffiliateRepo) SumCommissionsByLevel(referrerID uuid.UUID) (map[int]float64, error) { return nil, nil }
func (m *mockAffiliateRepo) ApproveCommission(id uuid.UUID) error { return nil }
func (m *mockAffiliateRepo) MarkPaid(id uuid.UUID) error          { return nil }
func (m *mockAffiliateRepo) GetDownlineFlat(rootID uuid.UUID, maxDepth int) ([]domain.MLMFlatNode, error) {
	return nil, nil
}
func (m *mockAffiliateRepo) GetUplineChain(userID uuid.UUID, maxDepth int) ([]domain.User, error) {
	return m.upline, nil
}
func (m *mockAffiliateRepo) CountDownline(userID uuid.UUID, level int) (int64, error) { return 0, nil }
func (m *mockAffiliateRepo) GetDirectDownlinePurchases(sponsorID uuid.UUID) ([]domain.DirectDownlinePurchase, error) {
	return nil, nil
}
func (m *mockAffiliateRepo) CountDirectDownlineConversions(sponsorID uuid.UUID) (int64, error) { return 0, nil }
func (m *mockAffiliateRepo) CountLegsWithMinRank(sponsorID uuid.UUID, minRank domain.MemberRank) (int, error) {
	return 0, nil
}
func (m *mockAffiliateRepo) CreateRankChangeLog(log *domain.RankChangeLog) error {
	m.rankLogs = append(m.rankLogs, log)
	return nil
}
func (m *mockAffiliateRepo) ListRankChangeLogs(userID uuid.UUID) ([]domain.RankChangeLog, error) { return nil, nil }

type mockUserRepo struct {
	users map[uuid.UUID]*domain.User
}

func (m *mockUserRepo) Create(user *domain.User) error                          { return nil }
func (m *mockUserRepo) FindByID(id uuid.UUID) (*domain.User, error)             { return m.users[id], nil }
func (m *mockUserRepo) FindByEmail(email string) (*domain.User, error)          { return nil, nil }
func (m *mockUserRepo) FindByPhone(phone string) (*domain.User, error)          { return nil, nil }
func (m *mockUserRepo) FindByReferralCode(code string) (*domain.User, error)    { return nil, nil }
func (m *mockUserRepo) FindByEmailVerifyToken(token string) (*domain.User, error) { return nil, nil }
func (m *mockUserRepo) FindByPasswordResetToken(token string) (*domain.User, error) { return nil, nil }
func (m *mockUserRepo) Update(user *domain.User) error                          { return nil }
func (m *mockUserRepo) Delete(id uuid.UUID) error                               { return nil }
func (m *mockUserRepo) List(role domain.UserRole, search string, page, pageSize int) ([]domain.User, int64, error) {
	return nil, 0, nil
}
func (m *mockUserRepo) Count() (int64, error)                             { return 0, nil }
func (m *mockUserRepo) CountByRole(role domain.UserRole) (int64, error)   { return 0, nil }
func (m *mockUserRepo) GetCRMSummary() (*domain.CRMSummary, error)        { return nil, nil }
func (m *mockUserRepo) GetUserCRMStats(userID uuid.UUID) (*domain.CRMUserStats, error) { return nil, nil }

type mockWalletSvc struct {
	credits map[uuid.UUID]float64
}

func (m *mockWalletSvc) GetOrCreateWallet(userID uuid.UUID) (*domain.Wallet, error) { return nil, nil }
func (m *mockWalletSvc) CreditPoints(userID uuid.UUID, amount float64, txType, refID, description string) (*domain.WalletTransaction, error) {
	m.credits[userID] += amount
	return &domain.WalletTransaction{Amount: amount}, nil
}
func (m *mockWalletSvc) CreditRM(userID uuid.UUID, amount float64, txType, refID, description string) (*domain.WalletTransaction, error) {
	return nil, nil
}
func (m *mockWalletSvc) DebitPoints(userID uuid.UUID, amount float64, txType, refID, description string) (*domain.WalletTransaction, error) {
	return nil, nil
}
func (m *mockWalletSvc) DebitRM(userID uuid.UUID, amount float64, txType, refID, description string) (*domain.WalletTransaction, error) {
	return nil, nil
}
func (m *mockWalletSvc) ListTransactions(userID uuid.UUID, page, pageSize int) ([]domain.WalletTransaction, int64, error) {
	return nil, 0, nil
}
func (m *mockWalletSvc) SumLifetimePointsEarned(userID uuid.UUID) (float64, error) { return 0, nil }
func (m *mockWalletSvc) GetWithdrawalInfo(userID uuid.UUID) (*domain.WithdrawalInfoResponse, error) {
	return nil, nil
}
func (m *mockWalletSvc) QuoteWithdrawal(amount float64) (*domain.WithdrawalQuoteResponse, error) { return nil, nil }
func (m *mockWalletSvc) CreateWithdrawal(userID uuid.UUID, req domain.CreateWithdrawalRequest) (*domain.WithdrawalRequest, error) {
	return nil, nil
}
func (m *mockWalletSvc) GetWithdrawal(id uuid.UUID, userID uuid.UUID, isAdmin bool) (*domain.WithdrawalRequest, error) {
	return nil, nil
}
func (m *mockWalletSvc) ListMyWithdrawals(userID uuid.UUID, status string, page, pageSize int) ([]domain.WithdrawalRequest, int64, error) {
	return nil, 0, nil
}
func (m *mockWalletSvc) ListAllWithdrawals(status string, page, pageSize int) ([]domain.WithdrawalRequest, int64, error) {
	return nil, 0, nil
}
func (m *mockWalletSvc) ApproveWithdrawal(id uuid.UUID, adminID uuid.UUID, remark string) (*domain.WithdrawalRequest, error) {
	return nil, nil
}
func (m *mockWalletSvc) RejectWithdrawal(id uuid.UUID, adminID uuid.UUID, remark string) (*domain.WithdrawalRequest, error) {
	return nil, nil
}
func (m *mockWalletSvc) ListWallets(page, pageSize int, search string) ([]domain.Wallet, int64, error) {
	return nil, 0, nil
}
func (m *mockWalletSvc) ListAllTransactions(walletID *uuid.UUID, page, pageSize int) ([]domain.WalletTransaction, int64, error) {
	return nil, 0, nil
}

func TestAamebaBonusCalculation_RB_And_DifferentialLRB(t *testing.T) {
	buyerID := uuid.New()
	s1ID := uuid.New()
	s2ID := uuid.New()
	s3ID := uuid.New()

	buyer := &domain.User{ID: buyerID, FirstName: "Maya", LastName: "Chong", IsActive: true, HighestRank: domain.RankMember}
	upline := []domain.User{
		{ID: s1ID, FirstName: "Jason", LastName: "Wong", IsActive: true, HighestRank: domain.RankOneStar},
		{ID: s2ID, FirstName: "Sarah", LastName: "Lim", IsActive: true, HighestRank: domain.RankTwoStar},
		{ID: s3ID, FirstName: "Elena", LastName: "Tan", IsActive: true, HighestRank: domain.RankThreeStar},
	}

	affRepo := &mockAffiliateRepo{
		settings: &domain.AffiliateSettings{
			MLMLevelRates:   "[15,15,15]",
			LeadershipRates: `{"S1":15,"S2":30,"S3":45}`,
			MLMDepth:        3,
			IsActive:        true,
		},
		upline: upline,
	}
	userRepo := &mockUserRepo{users: map[uuid.UUID]*domain.User{buyerID: buyer}}
	walletSvc := &mockWalletSvc{credits: map[uuid.UUID]float64{}}

	svc := NewAffiliateService(affRepo, userRepo, walletSvc, nil)

	order := &domain.Order{
		ID:          uuid.New(),
		OrderNo:     "AG-202611-TEST01",
		UserID:      buyerID,
		BuyerType:   "M",
		SubTotal:    120,
		TotalAmount: 120,
		TotalBV:     100, // 100 BV for crisp verification
	}

	if err := svc.ProcessOrderCommission(order); err != nil {
		t.Fatalf("ProcessOrderCommission failed: %v", err)
	}

	// Expected payouts on 100 BV (Aameba origin/dev exact math):
	// Buyer (M):   L0 Self RB (15%) = 15.00
	// Jason (S1):  L1 RB (15%) + LRB S1 (15% - 0%) = 15 + 15 = 30.00
	// Sarah (S2):  L2 RB (15%) + LRB S2 (30% - 15%) = 15 + 15 = 30.00
	// Elena (S3):  LRB S3 (45% - 30%) = 15.00
	if walletSvc.credits[buyerID] != 15 {
		t.Errorf("buyer expected 15 pts, got %.2f", walletSvc.credits[buyerID])
	}
	if walletSvc.credits[s1ID] != 30 {
		t.Errorf("S1 expected 30 pts (15 RB + 15 LRB), got %.2f", walletSvc.credits[s1ID])
	}
	if walletSvc.credits[s2ID] != 30 {
		t.Errorf("S2 expected 30 pts (15 RB + 15 LRB), got %.2f", walletSvc.credits[s2ID])
	}
	if walletSvc.credits[s3ID] != 15 {
		t.Errorf("S3 expected 15 pts (15 differential LRB), got %.2f", walletSvc.credits[s3ID])
	}
}
