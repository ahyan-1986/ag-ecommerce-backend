package service

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type walletService struct {
	walletRepo    domain.WalletRepository
	userRepo      domain.UserRepository
	affiliateRepo domain.AffiliateRepository
	db            *gorm.DB
}

func NewWalletService(
	walletRepo domain.WalletRepository,
	userRepo domain.UserRepository,
	affiliateRepo domain.AffiliateRepository,
	db *gorm.DB,
) domain.WalletService {
	return &walletService{
		walletRepo:    walletRepo,
		userRepo:      userRepo,
		affiliateRepo: affiliateRepo,
		db:            db,
	}
}

func (s *walletService) GetOrCreateWallet(userID uuid.UUID) (*domain.Wallet, error) {
	wallet, err := s.walletRepo.GetWalletByUserID(userID)
	if err != nil {
		return nil, err
	}
	if wallet != nil {
		return wallet, nil
	}

	newWallet := &domain.Wallet{
		ID:            uuid.New(),
		UserID:        userID,
		RMBalance:     0,
		PointsBalance: 0,
	}
	if err := s.walletRepo.CreateWallet(newWallet); err != nil {
		return nil, err
	}
	return newWallet, nil
}

func (s *walletService) CreditPoints(userID uuid.UUID, amount float64, txType, refID, description string) (*domain.WalletTransaction, error) {
	if amount <= 0 {
		return nil, errors.New("credit amount must be positive")
	}
	return s.processTransaction(userID, amount, 0, amount, "POINTS", txType, refID, description)
}

func (s *walletService) CreditRM(userID uuid.UUID, amount float64, txType, refID, description string) (*domain.WalletTransaction, error) {
	if amount <= 0 {
		return nil, errors.New("credit amount must be positive")
	}
	return s.processTransaction(userID, amount, amount, 0, "RM", txType, refID, description)
}

func (s *walletService) DebitPoints(userID uuid.UUID, amount float64, txType, refID, description string) (*domain.WalletTransaction, error) {
	if amount <= 0 {
		return nil, errors.New("debit amount must be positive")
	}
	return s.processTransaction(userID, -amount, 0, -amount, "POINTS", txType, refID, description)
}

func (s *walletService) DebitRM(userID uuid.UUID, amount float64, txType, refID, description string) (*domain.WalletTransaction, error) {
	if amount <= 0 {
		return nil, errors.New("debit amount must be positive")
	}
	return s.processTransaction(userID, -amount, -amount, 0, "RM", txType, refID, description)
}

func (s *walletService) processTransaction(userID uuid.UUID, txAmount, rmDelta, pointsDelta float64, balanceType, txType, refID, description string) (*domain.WalletTransaction, error) {
	var txRecord *domain.WalletTransaction

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var wallet domain.Wallet
		if err := tx.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				wallet = domain.Wallet{ID: uuid.New(), UserID: userID, RMBalance: 0, PointsBalance: 0}
				if err := tx.Create(&wallet).Error; err != nil {
					return fmt.Errorf("failed to create wallet: %w", err)
				}
			} else {
				return err
			}
		}

		if rmDelta < 0 && wallet.RMBalance+rmDelta < -0.0001 {
			return errors.New("insufficient RM balance")
		}
		if pointsDelta < 0 && wallet.PointsBalance+pointsDelta < -0.0001 {
			return errors.New("insufficient Loyalty Points balance")
		}

		if err := tx.Model(&domain.Wallet{}).Where("id = ?", wallet.ID).
			Updates(map[string]interface{}{
				"rm_balance":     gorm.Expr("rm_balance + ?", rmDelta),
				"points_balance": gorm.Expr("points_balance + ?", pointsDelta),
			}).Error; err != nil {
			return err
		}

		orderNo := ""
		if strings.HasPrefix(refID, "AG-") || strings.HasPrefix(refID, "HM-") {
			orderNo = refID
		}

		txRecord = &domain.WalletTransaction{
			WalletID:    wallet.ID,
			UserID:      userID,
			Amount:      txAmount,
			BalanceType: balanceType,
			Type:        txType,
			RefID:       refID,
			OrderNo:     orderNo,
			Description: description,
		}
		return tx.Create(txRecord).Error
	})

	if err != nil {
		return nil, err
	}
	return txRecord, nil
}

func (s *walletService) ListTransactions(userID uuid.UUID, page, pageSize int) ([]domain.WalletTransaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	return s.walletRepo.ListTransactions(userID, page, pageSize)
}

func (s *walletService) SumLifetimePointsEarned(userID uuid.UUID) (float64, error) {
	return s.walletRepo.SumLifetimePointsEarned(userID)
}

// ── Redeem Rewards / Withdrawals (Aameba origin/dev flow) ───────────────────

func (s *walletService) getWithdrawalConfig() (minAmt float64, feeRate float64) {
	minAmt = 50.0
	feeRate = 0.0
	if s.affiliateRepo != nil {
		if cfg, err := s.affiliateRepo.GetSettings(); err == nil && cfg != nil {
			if cfg.MinPayoutAmt > 0 {
				minAmt = cfg.MinPayoutAmt
			}
			feeRate = cfg.WithdrawalFeeRate
		}
	}
	return
}

func (s *walletService) GetWithdrawalInfo(userID uuid.UUID) (*domain.WithdrawalInfoResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	wallet, _ := s.GetOrCreateWallet(userID)
	avail := 0.0
	if wallet != nil {
		avail = math.Round(wallet.PointsBalance*100) / 100
	}
	minAmt, feeRate := s.getWithdrawalConfig()

	eligible := true
	reason := ""
	if strings.TrimSpace(user.BankName) == "" || strings.TrimSpace(user.BankAccountNo) == "" || strings.TrimSpace(user.PayeeName) == "" {
		eligible = false
		reason = "Please complete your Bank Account & Payee Details in My Profile before redeeming rewards."
	}

	return &domain.WithdrawalInfoResponse{
		Eligible:         eligible,
		IneligibleReason: reason,
		AvailablePoints:  avail,
		MinAmount:        minAmt,
		FeeRate:          feeRate,
		BankName:         user.BankName,
		BankAccountNo:    user.BankAccountNo,
		PayeeName:        user.PayeeName,
		PayeeIDNo:        user.PayeeIDNo,
	}, nil
}

func (s *walletService) QuoteWithdrawal(amount float64) (*domain.WithdrawalQuoteResponse, error) {
	_, feeRate := s.getWithdrawalConfig()
	fee := math.Round((amount*(feeRate/100.0))*100) / 100
	net := math.Round((amount-fee)*100) / 100
	if net < 0 {
		net = 0
	}
	return &domain.WithdrawalQuoteResponse{
		Amount:    math.Round(amount*100) / 100,
		FeeRate:   feeRate,
		FeeAmount: fee,
		NetAmount: net,
		Currency:  "MYR",
	}, nil
}

func (s *walletService) CreateWithdrawal(userID uuid.UUID, req domain.CreateWithdrawalRequest) (*domain.WithdrawalRequest, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Verify password (Aameba origin/dev requires password confirmation)
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("incorrect account password")
	}

	if strings.TrimSpace(user.BankName) == "" || strings.TrimSpace(user.BankAccountNo) == "" || strings.TrimSpace(user.PayeeName) == "" {
		return nil, errors.New("bank account details incomplete; please update My Profile first")
	}

	minAmt, feeRate := s.getWithdrawalConfig()
	if req.Amount < minAmt {
		return nil, fmt.Errorf("minimum redemption amount is %.2f points", minAmt)
	}

	fee := math.Round((req.Amount*(feeRate/100.0))*100) / 100
	net := math.Round((req.Amount-fee)*100) / 100
	receiptNo := fmt.Sprintf("WD-%s-%s", time.Now().Format("200601"), strings.ToUpper(uuid.New().String()[:6]))

	// Debit points immediately
	desc := fmt.Sprintf("Reward Redemption Request · %s", receiptNo)
	if _, err := s.DebitPoints(userID, req.Amount, "withdrawal", receiptNo, desc); err != nil {
		return nil, err
	}

	w := &domain.WithdrawalRequest{
		ReceiptNo:     receiptNo,
		UserID:        userID,
		Amount:        math.Round(req.Amount*100) / 100,
		FeeAmount:     fee,
		NetAmount:     net,
		Currency:      "MYR",
		BankName:      user.BankName,
		BankAccountNo: user.BankAccountNo,
		PayeeName:     user.PayeeName,
		PayeeIDNo:     user.PayeeIDNo,
		Status:        "pending",
	}
	if err := s.walletRepo.CreateWithdrawal(w); err != nil {
		// Refund if insert failed
		_, _ = s.CreditPoints(userID, req.Amount, "withdrawal_refund", receiptNo, "Refund failed withdrawal "+receiptNo)
		return nil, err
	}
	return w, nil
}

func (s *walletService) GetWithdrawal(id uuid.UUID, userID uuid.UUID, isAdmin bool) (*domain.WithdrawalRequest, error) {
	w, err := s.walletRepo.GetWithdrawalByID(id)
	if err != nil {
		return nil, err
	}
	if !isAdmin && w.UserID != userID {
		return nil, errors.New("access denied")
	}
	return w, nil
}

func (s *walletService) ListMyWithdrawals(userID uuid.UUID, status string, page, pageSize int) ([]domain.WithdrawalRequest, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	return s.walletRepo.ListWithdrawals(&userID, status, page, pageSize)
}

func (s *walletService) ListAllWithdrawals(status string, page, pageSize int) ([]domain.WithdrawalRequest, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.walletRepo.ListWithdrawals(nil, status, page, pageSize)
}

func (s *walletService) ApproveWithdrawal(id uuid.UUID, adminID uuid.UUID, remark string) (*domain.WithdrawalRequest, error) {
	w, err := s.walletRepo.GetWithdrawalByID(id)
	if err != nil {
		return nil, err
	}
	if w.Status != "pending" {
		return nil, errors.New("withdrawal request is already processed")
	}
	now := time.Now()
	w.Status = "approved"
	w.AdminRemark = remark
	w.ReviewedBy = &adminID
	w.ReviewedAt = &now
	return w, s.walletRepo.UpdateWithdrawal(w)
}

func (s *walletService) RejectWithdrawal(id uuid.UUID, adminID uuid.UUID, remark string) (*domain.WithdrawalRequest, error) {
	w, err := s.walletRepo.GetWithdrawalByID(id)
	if err != nil {
		return nil, err
	}
	if w.Status != "pending" {
		return nil, errors.New("withdrawal request is already processed")
	}
	now := time.Now()
	w.Status = "rejected"
	w.AdminRemark = remark
	w.ReviewedBy = &adminID
	w.ReviewedAt = &now
	if err := s.walletRepo.UpdateWithdrawal(w); err != nil {
		return nil, err
	}
	// Refund points back to member wallet
	desc := fmt.Sprintf("Redemption Refunded (%s) · %s", w.ReceiptNo, remark)
	_, _ = s.CreditPoints(w.UserID, w.Amount, "withdrawal_refund", w.ReceiptNo, desc)
	return w, nil
}

func (s *walletService) ListWallets(page, pageSize int, search string) ([]domain.Wallet, int64, error) {
	return s.walletRepo.ListWallets(page, pageSize, search)
}

func (s *walletService) ListAllTransactions(walletID *uuid.UUID, page, pageSize int) ([]domain.WalletTransaction, int64, error) {
	return s.walletRepo.ListAllTransactions(walletID, page, pageSize)
}
