package repository

import (
	"time"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type voucherRepository struct{ db *gorm.DB }

func NewVoucherRepository(db *gorm.DB) domain.VoucherRepository {
	return &voucherRepository{db: db}
}

func (r *voucherRepository) Create(v *domain.Voucher) error {
	return r.db.Create(v).Error
}

func (r *voucherRepository) FindByID(id uuid.UUID) (*domain.Voucher, error) {
	var v domain.Voucher
	err := r.db.First(&v, "id = ?", id).Error
	return &v, err
}

func (r *voucherRepository) FindByCode(code string) (*domain.Voucher, error) {
	var v domain.Voucher
	err := r.db.First(&v, "UPPER(code) = UPPER(?)", code).Error
	return &v, err
}

func (r *voucherRepository) Update(v *domain.Voucher) error {
	return r.db.Save(v).Error
}

func (r *voucherRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Voucher{}, "id = ?", id).Error
}

func (r *voucherRepository) List(active *bool, page, pageSize int) ([]domain.Voucher, int64, error) {
	q := r.db.Model(&domain.Voucher{})
	if active != nil {
		q = q.Where("active = ?", *active)
	}
	var total int64
	q.Count(&total)
	var items []domain.Voucher
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *voucherRepository) IncrementUsed(id uuid.UUID) error {
	return r.db.Model(&domain.Voucher{}).
		Where("id = ?", id).
		UpdateColumn("used_qty", gorm.Expr("used_qty + 1")).Error
}

func (r *voucherRepository) CreateRedemption(rd *domain.VoucherRedemption) error {
	return r.db.Create(rd).Error
}

func (r *voucherRepository) CountRedemptionsByUser(voucherID, userID uuid.UUID) (int, error) {
	var c int64
	err := r.db.Model(&domain.VoucherRedemption{}).
		Where("voucher_id = ? AND user_id = ?", voucherID, userID).
		Count(&c).Error
	return int(c), err
}

func (r *voucherRepository) ListRedemptions(voucherID *uuid.UUID, page, pageSize int) ([]domain.VoucherRedemption, int64, error) {
	q := r.db.Model(&domain.VoucherRedemption{}).Preload("Voucher")
	if voucherID != nil {
		q = q.Where("voucher_id = ?", *voucherID)
	}
	var total int64
	q.Count(&total)
	var items []domain.VoucherRedemption
	offset := (page - 1) * pageSize
	err := q.Order("redeemed_at DESC").Offset(offset).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *voucherRepository) CreateUserVoucher(uv *domain.UserVoucher) error {
	return r.db.Create(uv).Error
}

func (r *voucherRepository) ListUserVouchers(userID uuid.UUID) ([]domain.UserVoucher, error) {
	var items []domain.UserVoucher
	err := r.db.Preload("Voucher").
		Where("user_id = ? AND used_at IS NULL", userID).
		Order("claimed_at DESC").
		Find(&items).Error
	return items, err
}

func (r *voucherRepository) FindUnusedUserVoucher(voucherID, userID uuid.UUID) (*domain.UserVoucher, error) {
	var uv domain.UserVoucher
	err := r.db.Where("voucher_id = ? AND user_id = ? AND used_at IS NULL", voucherID, userID).First(&uv).Error
	return &uv, err
}

func (r *voucherRepository) MarkUserVoucherUsed(voucherID, userID, orderID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&domain.UserVoucher{}).
		Where("voucher_id = ? AND user_id = ? AND used_at IS NULL", voucherID, userID).
		Updates(map[string]interface{}{"used_at": &now, "order_id": &orderID}).Error
}

func (r *voucherRepository) ListClaimableVouchers() ([]domain.Voucher, error) {
	var items []domain.Voucher
	now := time.Now()
	err := r.db.Where("active = ? AND is_claimable = ? AND (expires_at IS NULL OR expires_at > ?)", true, true, now).
		Order("created_at DESC").
		Find(&items).Error
	return items, err
}
