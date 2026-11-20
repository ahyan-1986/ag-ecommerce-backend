package repository

import (
	"fmt"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(u *domain.User) error {
	return r.db.Create(u).Error
}

func (r *userRepository) FindByID(id uuid.UUID) (*domain.User, error) {
	var u domain.User
	if err := r.db.First(&u, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var u domain.User
	if err := r.db.First(&u, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByPhone(phone string) (*domain.User, error) {
	var u domain.User
	if err := r.db.First(&u, "phone = ?", phone).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByReferralCode(code string) (*domain.User, error) {
	var u domain.User
	if err := r.db.First(&u, "referral_code = ?", code).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByEmailVerifyToken(token string) (*domain.User, error) {
	var u domain.User
	if err := r.db.First(&u, "email_verify_token = ?", token).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByPasswordResetToken(token string) (*domain.User, error) {
	var u domain.User
	if err := r.db.First(&u, "password_reset_token = ?", token).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Update(u *domain.User) error {
	return r.db.Save(u).Error
}

func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.User{}, "id = ?", id).Error
}

func (r *userRepository) List(role domain.UserRole, search string, page, pageSize int) ([]domain.User, int64, error) {
	q := r.db.Model(&domain.User{})
	if role != "" {
		q = q.Where("role = ?", role)
	}
	if search != "" {
		like := fmt.Sprintf("%%%s%%", search)
		q = q.Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ? OR phone ILIKE ?", like, like, like, like)
	}
	var total int64
	q.Count(&total)
	var users []domain.User
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error
	return users, total, err
}

func (r *userRepository) Count() (int64, error) {
	var c int64
	err := r.db.Model(&domain.User{}).Count(&c).Error
	return c, err
}

func (r *userRepository) CountByRole(role domain.UserRole) (int64, error) {
	var c int64
	err := r.db.Model(&domain.User{}).Where("role = ?", role).Count(&c).Error
	return c, err
}

func (r *userRepository) GetCRMSummary() (*domain.CRMSummary, error) {
	var s domain.CRMSummary
	r.db.Raw(`SELECT COUNT(*) FROM users WHERE role = 'customer' AND deleted_at IS NULL`).Scan(&s.TotalCustomers)
	r.db.Raw(`SELECT COUNT(*) FROM users WHERE role = 'customer' AND is_active = true AND deleted_at IS NULL`).Scan(&s.ActiveCustomers)
	r.db.Raw(`SELECT COUNT(*) FROM users WHERE role = 'customer' AND created_at >= date_trunc('month', now()) AND deleted_at IS NULL`).Scan(&s.NewThisMonth)
	r.db.Raw(`SELECT COUNT(*) FROM users WHERE role = 'customer' AND referral_code != '' AND deleted_at IS NULL`).Scan(&s.WithReferralCode)
	r.db.Raw(`SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE status NOT IN ('cancelled','refunded')`).Scan(&s.TotalRevenue)
	var orderCount int64
	r.db.Raw(`SELECT COUNT(*) FROM orders WHERE status NOT IN ('cancelled','refunded')`).Scan(&orderCount)
	if orderCount > 0 {
		s.AvgOrderValue = s.TotalRevenue / float64(orderCount)
	}
	return &s, nil
}

func (r *userRepository) GetUserCRMStats(userID uuid.UUID) (*domain.CRMUserStats, error) {
	stats := &domain.CRMUserStats{UserID: userID.String()}
	r.db.Raw(`SELECT COUNT(*), COALESCE(SUM(total_amount),0) FROM orders WHERE user_id = ? AND status NOT IN ('cancelled','refunded')`, userID).Row().Scan(&stats.TotalOrders, &stats.TotalSpend)
	var lastOrder struct{ LastAt string }
	r.db.Raw(`SELECT TO_CHAR(MAX(created_at), 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS last_at FROM orders WHERE user_id = ?`, userID).Scan(&lastOrder)
	stats.LastOrderAt = lastOrder.LastAt
	r.db.Raw(`SELECT COUNT(*) FROM users WHERE referred_by_id = ? AND deleted_at IS NULL`, userID).Scan(&stats.DownlineCount)
	r.db.Raw(`SELECT COALESCE(SUM(commission_amt),0) FROM affiliate_commissions WHERE referrer_id = ? AND status != 'rejected'`, userID).Scan(&stats.CommissionEarned)
	return stats, nil
}
