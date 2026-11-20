package repository

import (
	"time"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type affiliateRepository struct{ db *gorm.DB }

func NewAffiliateRepository(db *gorm.DB) domain.AffiliateRepository {
	return &affiliateRepository{db: db}
}

func (r *affiliateRepository) GetSettings() (*domain.AffiliateSettings, error) {
	var s domain.AffiliateSettings
	err := r.db.First(&s).Error
	if err == gorm.ErrRecordNotFound {
		s = domain.AffiliateSettings{
			CommissionRate:      15.0,
			MLMLevelRates:       "[15,15,15]",
			LeadershipRates:     `{"S1":15,"S2":30,"S3":45}`,
			MLMDepth:            3,
			OneStarMinDownlines: 10,
			OneStarMinSpend:     8000,
			TwoStarMinS1Legs:    5,
			ThreeStarMinS2Legs:  5,
			MinPayoutAmt:        50.0,
			WithdrawalFeeRate:   0,
			IsActive:            true,
		}
		r.db.Create(&s)
		return &s, nil
	}
	return &s, err
}

func (r *affiliateRepository) UpdateSettings(s *domain.AffiliateSettings) error {
	return r.db.Save(s).Error
}

func (r *affiliateRepository) CreateCommission(c *domain.AffiliateCommission) error {
	return r.db.Create(c).Error
}

func (r *affiliateRepository) FindCommissionByOrder(orderID uuid.UUID) (*domain.AffiliateCommission, error) {
	var c domain.AffiliateCommission
	err := r.db.Where("order_id = ?", orderID).First(&c).Error
	return &c, err
}

func (r *affiliateRepository) GetCommissionByID(id uuid.UUID) (*domain.AffiliateCommission, error) {
	var c domain.AffiliateCommission
	err := r.db.Where("id = ?", id).First(&c).Error
	return &c, err
}

func (r *affiliateRepository) ListCommissions(referrerID *uuid.UUID, status string, page, pageSize int) ([]domain.AffiliateCommission, int64, error) {
	q := r.db.Model(&domain.AffiliateCommission{}).Preload("Referrer").Preload("Referee")
	if referrerID != nil {
		q = q.Where("referrer_id = ?", *referrerID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	q.Count(&total)
	var items []domain.AffiliateCommission
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *affiliateRepository) SumCommissions(referrerID uuid.UUID, status string) (float64, error) {
	var total float64
	q := r.db.Model(&domain.AffiliateCommission{}).
		Where("referrer_id = ?", referrerID).
		Select("COALESCE(SUM(commission_amt), 0)")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Scan(&total).Error
	return total, err
}

func (r *affiliateRepository) SumCommissionsByLevel(referrerID uuid.UUID) (map[int]float64, error) {
	type result struct {
		Level           int
		TotalCommission float64
	}
	var rows []result
	err := r.db.Model(&domain.AffiliateCommission{}).
		Select("level, COALESCE(SUM(commission_amt), 0) as total_commission").
		Where("referrer_id = ?", referrerID).
		Group("level").
		Order("level").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	m := make(map[int]float64)
	for _, row := range rows {
		m[row.Level] = row.TotalCommission
	}
	return m, nil
}

func (r *affiliateRepository) ApproveCommission(id uuid.UUID) error {
	return r.db.Model(&domain.AffiliateCommission{}).
		Where("id = ?", id).
		Update("status", "approved").Error
}

func (r *affiliateRepository) MarkPaid(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&domain.AffiliateCommission{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"status": "paid", "paid_at": &now}).Error
}

func (r *affiliateRepository) GetDownlineFlat(rootID uuid.UUID, maxDepth int) ([]domain.MLMFlatNode, error) {
	type row struct {
		UserID       string
		MemberCode   string
		FirstName    string
		LastName     string
		Email        string
		ReferralCode string
		MemberRank   string
		Level        int
		ParentID     *string
		JoinedAt     time.Time
	}

	query := `
		WITH RECURSIVE downline AS (
			SELECT
				u.id::text        AS user_id,
				u.member_code,
				u.first_name,
				u.last_name,
				u.email,
				u.referral_code,
				u.member_rank,
				1                 AS level,
				u.referred_by_id::text AS parent_id,
				u.created_at      AS joined_at
			FROM users u
			WHERE u.referred_by_id = ?
			  AND u.deleted_at IS NULL

			UNION ALL

			SELECT
				u.id::text,
				u.member_code,
				u.first_name,
				u.last_name,
				u.email,
				u.referral_code,
				u.member_rank,
				d.level + 1,
				u.referred_by_id::text,
				u.created_at
			FROM users u
			INNER JOIN downline d ON u.referred_by_id::text = d.user_id
			WHERE d.level < ?
			  AND u.deleted_at IS NULL
		)
		SELECT * FROM downline ORDER BY level, joined_at
	`

	var rows []row
	if err := r.db.Raw(query, rootID.String(), maxDepth).Scan(&rows).Error; err != nil {
		return nil, err
	}

	nodes := make([]domain.MLMFlatNode, 0, len(rows))
	for _, rw := range rows {
		uid, err := uuid.Parse(rw.UserID)
		if err != nil {
			continue
		}
		rank := domain.MemberRank(rw.MemberRank)
		if rank == "" {
			rank = domain.RankMember
		}
		node := domain.MLMFlatNode{
			UserID:       uid,
			MemberCode:   rw.MemberCode,
			FirstName:    rw.FirstName,
			LastName:     rw.LastName,
			Email:        rw.Email,
			ReferralCode: rw.ReferralCode,
			MemberRank:   rank,
			Level:        rw.Level,
			JoinedAt:     rw.JoinedAt,
		}
		if rw.ParentID != nil {
			pid, err2 := uuid.Parse(*rw.ParentID)
			if err2 == nil {
				node.ParentID = &pid
			}
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (r *affiliateRepository) GetUplineChain(userID uuid.UUID, maxDepth int) ([]domain.User, error) {
	query := `
		WITH RECURSIVE upline AS (
			SELECT u.*, 1 AS depth
			FROM users u
			INNER JOIN users me ON me.referred_by_id = u.id
			WHERE me.id = ?
			  AND u.deleted_at IS NULL

			UNION ALL

			SELECT u.*, up.depth + 1
			FROM users u
			INNER JOIN upline up ON up.referred_by_id = u.id
			WHERE up.depth < ?
			  AND u.deleted_at IS NULL
		)
		SELECT * FROM upline ORDER BY depth
	`
	var users []domain.User
	err := r.db.Raw(query, userID.String(), maxDepth).Scan(&users).Error
	return users, err
}

func (r *affiliateRepository) CountDownline(userID uuid.UUID, level int) (int64, error) {
	if level == 1 {
		var count int64
		err := r.db.Model(&domain.User{}).
			Where("referred_by_id = ? AND deleted_at IS NULL", userID).
			Count(&count).Error
		return count, err
	}
	nodes, err := r.GetDownlineFlat(userID, 10)
	if err != nil {
		return 0, err
	}
	if level == 0 {
		return int64(len(nodes)), nil
	}
	var count int64
	for _, n := range nodes {
		if n.Level == level {
			count++
		}
	}
	return count, nil
}

func (r *affiliateRepository) GetDirectDownlinePurchases(sponsorID uuid.UUID) ([]domain.DirectDownlinePurchase, error) {
	type row struct {
		UserID     string
		MemberCode string
		FirstName  string
		LastName   string
		MemberRank string
		TotalSpend float64
		OrderCount int64
	}
	query := `
		SELECT
			u.id::text AS user_id,
			u.member_code,
			u.first_name,
			u.last_name,
			u.member_rank,
			COALESCE(SUM(CASE WHEN o.status IN ('paid','processing','shipped','delivered') AND o.buyer_type = 'M' THEN o.total_amount ELSE 0 END), 0) AS total_spend,
			COUNT(CASE WHEN o.status IN ('paid','processing','shipped','delivered') AND o.buyer_type = 'M' THEN 1 END) AS order_count
		FROM users u
		LEFT JOIN orders o ON o.user_id = u.id
		WHERE u.referred_by_id = ?
		  AND u.deleted_at IS NULL
		GROUP BY u.id, u.member_code, u.first_name, u.last_name, u.member_rank
		ORDER BY total_spend DESC
	`
	var rows []row
	if err := r.db.Raw(query, sponsorID.String()).Scan(&rows).Error; err != nil {
		return nil, err
	}
	res := make([]domain.DirectDownlinePurchase, 0, len(rows))
	for _, rw := range rows {
		uid, err := uuid.Parse(rw.UserID)
		if err != nil {
			continue
		}
		rank := domain.MemberRank(rw.MemberRank)
		if rank == "" {
			rank = domain.RankMember
		}
		res = append(res, domain.DirectDownlinePurchase{
			UserID:     uid,
			MemberCode: rw.MemberCode,
			Name:       rw.FirstName + " " + rw.LastName,
			MemberRank: rank,
			TotalSpend: rw.TotalSpend,
			OrderCount: rw.OrderCount,
		})
	}
	return res, nil
}

func (r *affiliateRepository) CountDirectDownlineConversions(sponsorID uuid.UUID) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(o.id)
		FROM orders o
		INNER JOIN users u ON o.user_id = u.id
		WHERE u.referred_by_id = ?
		  AND u.deleted_at IS NULL
		  AND o.status IN ('paid','processing','shipped','delivered')
	`
	err := r.db.Raw(query, sponsorID.String()).Scan(&count).Error
	return count, err
}

// CountLegsWithMinRank counts how many distinct direct-downline legs under sponsorID
// contain at least one member holding highest_rank >= minRank (matching Aameba qs1Line / qs2Line).
func (r *affiliateRepository) CountLegsWithMinRank(sponsorID uuid.UUID, minRank domain.MemberRank) (int, error) {
	var directs []domain.User
	if err := r.db.Where("referred_by_id = ? AND deleted_at IS NULL", sponsorID).Find(&directs).Error; err != nil {
		return 0, err
	}
	minOrd := minRank.Ordinal()
	qualifiedLegs := 0
	for _, legRoot := range directs {
		if legRoot.HighestRank.Ordinal() >= minOrd || legRoot.MemberRank.Ordinal() >= minOrd {
			qualifiedLegs++
			continue
		}
		nodes, err := r.GetDownlineFlat(legRoot.ID, 10)
		if err != nil {
			continue
		}
		for _, n := range nodes {
			if n.MemberRank.Ordinal() >= minOrd {
				qualifiedLegs++
				break
			}
		}
	}
	return qualifiedLegs, nil
}

func (r *affiliateRepository) CreateRankChangeLog(l *domain.RankChangeLog) error {
	return r.db.Create(l).Error
}

func (r *affiliateRepository) ListRankChangeLogs(userID uuid.UUID) ([]domain.RankChangeLog, error) {
	var logs []domain.RankChangeLog
	err := r.db.Where("user_id = ?", userID).Order("achieved_at DESC").Find(&logs).Error
	return logs, err
}
