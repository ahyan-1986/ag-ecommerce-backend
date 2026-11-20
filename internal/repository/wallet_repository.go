package repository

import (
	"errors"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) domain.WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) GetWalletByUserID(userID uuid.UUID) (*domain.Wallet, error) {
	var wallet domain.Wallet
	err := r.db.Where("user_id = ?", userID).First(&wallet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &wallet, err
}

func (r *walletRepository) CreateWallet(wallet *domain.Wallet) error {
	return r.db.Create(wallet).Error
}

func (r *walletRepository) CreateTransaction(tx *domain.WalletTransaction) error {
	return r.db.Create(tx).Error
}

func (r *walletRepository) UpdateWalletBalances(walletID uuid.UUID, rmDelta, pointsDelta float64) error {
	return r.db.Model(&domain.Wallet{}).
		Where("id = ?", walletID).
		Updates(map[string]interface{}{
			"rm_balance":     gorm.Expr("rm_balance + ?", rmDelta),
			"points_balance": gorm.Expr("points_balance + ?", pointsDelta),
		}).Error
}

func (r *walletRepository) ListTransactions(userID uuid.UUID, page, pageSize int) ([]domain.WalletTransaction, int64, error) {
	var txs []domain.WalletTransaction
	var total int64

	query := r.db.Model(&domain.WalletTransaction{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&txs).Error; err != nil {
		return nil, 0, err
	}

	return txs, total, nil
}

func (r *walletRepository) SumLifetimePointsEarned(userID uuid.UUID) (float64, error) {
	var total float64
	err := r.db.Model(&domain.WalletTransaction{}).
		Where("user_id = ? AND balance_type = 'POINTS' AND amount > 0", userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

// ── Withdrawals (Redeem Rewards) ─────────────────────────────────────────────

func (r *walletRepository) CreateWithdrawal(w *domain.WithdrawalRequest) error {
	return r.db.Create(w).Error
}

func (r *walletRepository) GetWithdrawalByID(id uuid.UUID) (*domain.WithdrawalRequest, error) {
	var w domain.WithdrawalRequest
	err := r.db.Preload("User").First(&w, "id = ?", id).Error
	return &w, err
}

func (r *walletRepository) UpdateWithdrawal(w *domain.WithdrawalRequest) error {
	return r.db.Save(w).Error
}

func (r *walletRepository) ListWithdrawals(userID *uuid.UUID, status string, page, pageSize int) ([]domain.WithdrawalRequest, int64, error) {
	q := r.db.Model(&domain.WithdrawalRequest{}).Preload("User")
	if userID != nil {
		q = q.Where("user_id = ?", *userID)
	}
	if status != "" && status != "all" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []domain.WithdrawalRequest
	offset := (page - 1) * pageSize
	if err := q.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ── Admin ─────────────────────────────────────────────────────────────

func (r *walletRepository) ListWallets(page, pageSize int, search string) ([]domain.Wallet, int64, error) {
	var users []domain.User
	var total int64

	query := r.db.Model(&domain.User{})
	if search != "" {
		searchQuery := "%" + search + "%"
		query = query.Where("email ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?", searchQuery, searchQuery, searchQuery)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	var userIDs []uuid.UUID
	for _, u := range users {
		userIDs = append(userIDs, u.ID)
	}

	var existingWallets []domain.Wallet
	if len(userIDs) > 0 {
		r.db.Where("user_id IN ?", userIDs).Find(&existingWallets)
	}

	walletMap := make(map[uuid.UUID]domain.Wallet)
	for _, w := range existingWallets {
		walletMap[w.UserID] = w
	}

	var result []domain.Wallet
	for _, u := range users {
		userCopy := u
		if w, ok := walletMap[u.ID]; ok {
			w.User = &userCopy
			result = append(result, w)
		} else {
			result = append(result, domain.Wallet{
				UserID:        u.ID,
				User:          &userCopy,
				RMBalance:     0,
				PointsBalance: 0,
			})
		}
	}

	return result, total, nil
}

func (r *walletRepository) ListAllTransactions(walletID *uuid.UUID, page, pageSize int) ([]domain.WalletTransaction, int64, error) {
	var txs []domain.WalletTransaction
	var total int64

	query := r.db.Model(&domain.WalletTransaction{})
	if walletID != nil {
		query = query.Where("wallet_id = ?", *walletID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&txs).Error; err != nil {
		return nil, 0, err
	}

	return txs, total, nil
}
