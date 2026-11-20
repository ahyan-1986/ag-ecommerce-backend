package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	BonusTypeRB  = "RB"  // Rebate Bonus (L0 self 15%, L1 sponsor 15%, L2 upline 15%)
	BonusTypeLRB = "LRB" // Leadership Rebate Bonus (differential: S1=15%, S2=30%, S3=45%)
)

// AffiliateCommission records every MLM rebate (RB) or leadership (LRB) bonus earned.
// Level 0 = self rebate, Level 1 = direct sponsor, Level 2 = 2nd upline.
type AffiliateCommission struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ReferrerID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"referrerId"` // earner user ID
	RefereeID      *uuid.UUID `gorm:"type:uuid;index" json:"refereeId,omitempty"` // giver user ID (nil for unregistered guest)
	OrderID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"orderId"`
	OrderNo        string     `gorm:"type:varchar(30);not null" json:"orderNo"`
	BonusType      string     `gorm:"type:varchar(10);not null;default:'RB'" json:"bonusType"` // RB | LRB
	GiverName      string     `gorm:"type:varchar(150);default:''" json:"giverName"`
	EarnerRank     MemberRank `gorm:"type:varchar(10);default:'M'" json:"earnerRank"`
	SaleAmount     float64    `gorm:"type:numeric(12,2);not null" json:"saleAmount"`
	OrderBV        float64    `gorm:"type:numeric(12,2);not null;default:0" json:"orderBv"`
	CommissionRate float64    `gorm:"type:numeric(5,2);not null" json:"commissionRate"` // %
	CommissionAmt  float64    `gorm:"type:numeric(12,2);not null" json:"commissionAmt"`
	Currency       string     `gorm:"type:varchar(10);not null;default:'MYR'" json:"currency"`
	Status         string     `gorm:"type:varchar(20);not null;default:'paid'" json:"status"` // pending, approved, paid
	Level          int        `gorm:"not null;default:1" json:"level"`                        // 0 = self, 1 = direct, 2 = L2
	Remark         string     `gorm:"type:varchar(255);default:''" json:"remark"`
	PaidAt         *time.Time `gorm:"type:timestamptz" json:"paidAt,omitempty"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"createdAt"`

	Referrer *User `gorm:"foreignKey:ReferrerID" json:"referrer,omitempty"`
	Referee  *User `gorm:"foreignKey:RefereeID" json:"referee,omitempty"`
}

func (AffiliateCommission) TableName() string { return "affiliate_commissions" }

// RankChangeLog audits every member rank promotion (matching Aameba MemberRankTrace).
type RankChangeLog struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"userId"`
	FromRank   MemberRank `gorm:"type:varchar(10);not null" json:"fromRank"`
	ToRank     MemberRank `gorm:"type:varchar(10);not null" json:"toRank"`
	Source     string     `gorm:"type:varchar(40);not null" json:"source"` // PAYMENT_PROMOTION, DAILY_CALC, ADMIN_MANUAL
	DoneBy     string     `gorm:"type:varchar(120);default:''" json:"doneBy"`
	AchievedAt time.Time  `gorm:"autoCreateTime" json:"achievedAt"`
}

func (RankChangeLog) TableName() string { return "rank_change_logs" }

// AffiliateSettings stores the global Aameba-aligned MLM & loyalty configuration.
type AffiliateSettings struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CommissionRate      float64   `gorm:"type:numeric(5,2);not null;default:15" json:"commissionRate"`
	MLMLevelRates       string    `gorm:"type:text;not null;default:'[15,15,15]'" json:"mlmLevelRates"`        // [L0 Self, L1 Sponsor, L2 Upline] % of BV
	LeadershipRates     string    `gorm:"type:text;not null;default:'{\"S1\":15,\"S2\":30,\"S3\":45}'" json:"leadershipRates"` // Differential % of BV
	MLMDepth            int       `gorm:"not null;default:3" json:"mlmDepth"`
	OneStarMinDownlines int       `gorm:"not null;default:10" json:"oneStarMinDownlines"`
	OneStarMinSpend     float64   `gorm:"type:numeric(12,2);not null;default:8000" json:"oneStarMinSpend"`
	TwoStarMinS1Legs    int       `gorm:"not null;default:5" json:"twoStarMinS1Legs"`
	ThreeStarMinS2Legs  int       `gorm:"not null;default:5" json:"threeStarMinS2Legs"`
	MinPayoutAmt        float64   `gorm:"type:numeric(10,2);not null;default:50" json:"minPayoutAmt"`
	WithdrawalFeeRate   float64   `gorm:"type:numeric(5,2);not null;default:0" json:"withdrawalFeeRate"` // % fee on withdrawal
	IsActive            bool      `gorm:"not null;default:true" json:"isActive"`
	UpdatedBy           uuid.UUID `gorm:"type:uuid" json:"updatedBy"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (AffiliateSettings) TableName() string { return "affiliate_settings" }

// ── Request / Response DTOs ───────────────────────────────────────────────────

type AffiliateStats struct {
	TotalEarned    float64 `json:"totalEarned"`
	PendingEarned  float64 `json:"pendingEarned"`
	TotalReferrals int64   `json:"totalReferrals"`
	MyReferralCode string  `json:"myReferralCode"`
}

type MLMNetworkStats struct {
	TotalDownline   int64           `json:"totalDownline"`
	DirectReferrals int64           `json:"directReferrals"`
	TotalEarned     float64         `json:"totalEarned"`
	PendingEarned   float64         `json:"pendingEarned"`
	EarnedByLevel   map[int]float64 `json:"earnedByLevel"`
	DownlineByLevel map[int]int64   `json:"downlineByLevel"`
	MyReferralCode  string          `json:"myReferralCode"`
	ReferralLink    string          `json:"referralLink"`
	LevelRates      []float64       `json:"levelRates"`
	LeadershipRates map[string]float64 `json:"leadershipRates"`
	TotalConversions int64          `json:"totalConversions"`
}

type MLMTreeNode struct {
	UserID       uuid.UUID      `json:"userId"`
	MemberCode   string         `json:"memberCode"`
	FirstName    string         `json:"firstName"`
	LastName     string         `json:"lastName"`
	Email        string         `json:"email"`
	ReferralCode string         `json:"referralCode"`
	MemberRank   MemberRank     `json:"memberRank"`
	Level        int            `json:"level"`
	JoinedAt     time.Time      `json:"joinedAt"`
	Children     []*MLMTreeNode `json:"children"`
	TotalOrders  int64          `json:"totalOrders"`
	TotalSpend   float64        `json:"totalSpend"`
	TotalEarned  float64        `json:"totalEarned"`
}

type SponsorInfo struct {
	HasSponsor   bool       `json:"hasSponsor"`
	SponsorID    uuid.UUID  `json:"sponsorId,omitempty"`
	MemberCode   string     `json:"memberCode,omitempty"`
	FirstName    string     `json:"firstName,omitempty"`
	LastName     string     `json:"lastName,omitempty"`
	Email        string     `json:"email,omitempty"`
	ReferralCode string     `json:"referralCode,omitempty"`
	MemberRank   MemberRank `json:"memberRank,omitempty"`
	JoinedAt     time.Time  `json:"joinedAt,omitempty"`
}

type MLMFlatNode struct {
	UserID       uuid.UUID
	MemberCode   string
	FirstName    string
	LastName     string
	Email        string
	ReferralCode string
	MemberRank   MemberRank
	Level        int
	ParentID     *uuid.UUID
	JoinedAt     time.Time
}

// RankProgress mirrors Aameba origin/dev's 4-rung rank ladder + live requirement bars.
type RankRung struct {
	Code    MemberRank `json:"code"`
	Label   string     `json:"label"`
	Earned  bool       `json:"earned"`
	Current bool       `json:"current"`
}

type RankRequirement struct {
	Metric  string  `json:"metric"` // SPENDING_DOWNLINES, COMBINED_SPEND, ONE_STAR_LEGS, TWO_STAR_LEGS
	Label   string  `json:"label"`
	Kind    string  `json:"kind"` // count | money
	Current float64 `json:"current"`
	Target  float64 `json:"target"`
	Percent int     `json:"percent"`
}

type RankProgressResponse struct {
	CurrentRank       MemberRank        `json:"currentRank"`
	CurrentRankLabel  string            `json:"currentRankLabel"`
	NextRank          *MemberRank       `json:"nextRank,omitempty"`
	NextRankLabel     string            `json:"nextRankLabel,omitempty"`
	Rungs             []RankRung        `json:"rungs"`
	Requirements      []RankRequirement `json:"requirements"`
	PendingSettlement bool              `json:"pendingSettlement"`
}

type MemberDashboardResponse struct {
	MemberCode      string               `json:"memberCode"`
	ReferralCode    string               `json:"referralCode"`
	MemberRank      MemberRank           `json:"memberRank"`
	MemberRankLabel string               `json:"memberRankLabel"`
	TotalReferrals  int64                `json:"totalReferrals"`
	TotalDownline   int64                `json:"totalDownline"`
	AvailablePoints float64              `json:"availablePoints"`
	LifetimePoints  float64              `json:"lifetimePoints"`
	RankProgress    RankProgressResponse `json:"rankProgress"`
}

type DirectDownlinePurchase struct {
	UserID     uuid.UUID  `json:"userId"`
	MemberCode string     `json:"memberCode"`
	Name       string     `json:"name"`
	MemberRank MemberRank `json:"memberRank"`
	TotalSpend float64    `json:"totalSpend"`
	OrderCount int64      `json:"orderCount"`
}

type MemberRankTraceResponse struct {
	UserID            uuid.UUID                `json:"userId"`
	MemberCode        string                   `json:"memberCode"`
	CurrentRank       MemberRank               `json:"currentRank"`
	RankChanges       []RankChangeLog          `json:"rankChanges"`
	DirectDownlines   []DirectDownlinePurchase `json:"directDownlines"`
	QualifiedCount    int                      `json:"qualifiedCount"`
	CombinedSpend     float64                  `json:"combinedSpend"`
	MinDownlines      int                      `json:"minDownlines"`
	MinCombinedSpend  float64                  `json:"minCombinedSpend"`
}

type UpdateAffiliateSettingsRequest struct {
	CommissionRate      float64            `json:"commissionRate"`
	MLMLevelRates       []float64          `json:"mlmLevelRates"`
	LeadershipRates     map[string]float64 `json:"leadershipRates"`
	MLMDepth            int                `json:"mlmDepth"`
	OneStarMinDownlines int                `json:"oneStarMinDownlines"`
	OneStarMinSpend     float64            `json:"oneStarMinSpend"`
	TwoStarMinS1Legs    int                `json:"twoStarMinS1Legs"`
	ThreeStarMinS2Legs  int                `json:"threeStarMinS2Legs"`
	MinPayoutAmt        float64            `json:"minPayoutAmt"`
	WithdrawalFeeRate   float64            `json:"withdrawalFeeRate"`
	IsActive            bool               `json:"isActive"`
}

// ── Interfaces ────────────────────────────────────────────────────────────────

type AffiliateRepository interface {
	GetSettings() (*AffiliateSettings, error)
	UpdateSettings(s *AffiliateSettings) error
	CreateCommission(c *AffiliateCommission) error
	FindCommissionByOrder(orderID uuid.UUID) (*AffiliateCommission, error)
	GetCommissionByID(id uuid.UUID) (*AffiliateCommission, error)
	ListCommissions(referrerID *uuid.UUID, status string, page, pageSize int) ([]AffiliateCommission, int64, error)
	SumCommissions(referrerID uuid.UUID, status string) (float64, error)
	SumCommissionsByLevel(referrerID uuid.UUID) (map[int]float64, error)
	ApproveCommission(id uuid.UUID) error
	MarkPaid(id uuid.UUID) error
	// MLM tree queries
	GetDownlineFlat(rootID uuid.UUID, maxDepth int) ([]MLMFlatNode, error)
	GetUplineChain(userID uuid.UUID, maxDepth int) ([]User, error)
	CountDownline(userID uuid.UUID, level int) (int64, error)
	// Rank & purchase summaries (Aameba origin/dev)
	GetDirectDownlinePurchases(sponsorID uuid.UUID) ([]DirectDownlinePurchase, error)
	CountDirectDownlineConversions(sponsorID uuid.UUID) (int64, error)
	CountLegsWithMinRank(sponsorID uuid.UUID, minRank MemberRank) (int, error)
	CreateRankChangeLog(log *RankChangeLog) error
	ListRankChangeLogs(userID uuid.UUID) ([]RankChangeLog, error)
}

type AffiliateService interface {
	GetSettings() (*AffiliateSettings, error)
	UpdateSettings(adminID uuid.UUID, req UpdateAffiliateSettingsRequest) (*AffiliateSettings, error)
	ProcessOrderCommission(order *Order) error
	ReleaseGuestHeldPoints(user *User) error
	EvaluateAndPromoteRanks(userID uuid.UUID, source string) error
	GetMyStats(userID uuid.UUID, referralCode string) (*AffiliateStats, error)
	GetMemberDashboard(userID uuid.UUID) (*MemberDashboardResponse, error)
	GetRankProgress(userID uuid.UUID) (*RankProgressResponse, error)
	GetRankTrace(userID uuid.UUID) (*MemberRankTraceResponse, error)
	SetMemberRankManual(userID uuid.UUID, newRank MemberRank, adminEmail string) error
	ListMyCommissions(userID uuid.UUID, page, pageSize int) ([]AffiliateCommission, int64, error)
	ListAllCommissions(status string, page, pageSize int) ([]AffiliateCommission, int64, error)
	ApproveCommission(id uuid.UUID) error
	MarkCommissionPaid(id uuid.UUID) error
	// MLM
	GetMyNetworkStats(userID uuid.UUID, referralCode string) (*MLMNetworkStats, error)
	GetMySponsor(userID uuid.UUID) (*SponsorInfo, error)
	GetMyNetworkTree(userID uuid.UUID) (*MLMTreeNode, error)
	GetAdminNetworkTree(rootUserID uuid.UUID) (*MLMTreeNode, error)
}
