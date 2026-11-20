package domain

import (
	"time"

	"github.com/google/uuid"
)

// Wallet represents a user's digital wallet, holding both RM and Loyalty Points.
type Wallet struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"userId"`
	RMBalance     float64   `gorm:"type:numeric(15,4);not null;default:0" json:"rmBalance"`
	PointsBalance float64   `gorm:"type:numeric(15,4);not null;default:0" json:"pointsBalance"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Wallet) TableName() string { return "wallets" }

// WalletTransaction records credits and debits to a wallet.
type WalletTransaction struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WalletID    uuid.UUID `gorm:"type:uuid;not null;index" json:"walletId"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	Amount      float64   `gorm:"type:numeric(15,4);not null" json:"amount"`               // Positive = credit, Negative = debit
	BalanceType string    `gorm:"type:varchar(20);not null;default:'POINTS'" json:"balanceType"` // "RM" or "POINTS"
	Type        string    `gorm:"type:varchar(50);not null" json:"type"`                  // "rebate_bonus", "leadership_bonus", "guest_rebate", "checkout_redeem", "withdrawal", "withdrawal_refund", "admin_adjustment"
	RefID       string    `gorm:"type:varchar(100)" json:"refId"`                         // Order No. or Withdrawal Receipt No.
	OrderNo     string    `gorm:"type:varchar(40);default:''" json:"orderNo"`
	Description string    `gorm:"type:varchar(255)" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`

	Wallet *Wallet `gorm:"foreignKey:WalletID" json:"-"`
}

func (WalletTransaction) TableName() string { return "wallet_transactions" }

// WithdrawalRequest mirrors Aameba origin/dev Redeem Rewards / Withdrawals flow.
type WithdrawalRequest struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ReceiptNo     string     `gorm:"type:varchar(40);uniqueIndex;not null" json:"receiptNo"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"userId"`
	Amount        float64    `gorm:"type:numeric(12,2);not null" json:"amount"`
	FeeAmount     float64    `gorm:"type:numeric(12,2);not null;default:0" json:"feeAmount"`
	NetAmount     float64    `gorm:"type:numeric(12,2);not null" json:"netAmount"`
	Currency      string     `gorm:"type:varchar(10);not null;default:'MYR'" json:"currency"`
	BankName      string     `gorm:"type:varchar(120);not null" json:"bankName"`
	BankAccountNo string     `gorm:"type:varchar(60);not null" json:"bankAccountNo"`
	PayeeName     string     `gorm:"type:varchar(150);not null" json:"payeeName"`
	PayeeIDNo     string     `gorm:"type:varchar(60);default:''" json:"payeeIdNo"`
	Status        string     `gorm:"type:varchar(20);not null;default:'pending'" json:"status"` // pending, approved, rejected
	AdminRemark   string     `gorm:"type:varchar(255);default:''" json:"adminRemark"`
	ReviewedBy    *uuid.UUID `gorm:"type:uuid" json:"reviewedBy,omitempty"`
	ReviewedAt    *time.Time `gorm:"type:timestamptz" json:"reviewedAt,omitempty"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (WithdrawalRequest) TableName() string { return "withdrawal_requests" }

// ── Withdrawal DTOs ──────────────────────────────────────────────────────────

type WithdrawalInfoResponse struct {
	Eligible         bool    `json:"eligible"`
	IneligibleReason string  `json:"ineligibleReason,omitempty"`
	AvailablePoints  float64 `json:"availablePoints"`
	MinAmount        float64 `json:"minAmount"`
	FeeRate          float64 `json:"feeRate"`
	BankName         string  `json:"bankName"`
	BankAccountNo    string  `json:"bankAccountNo"`
	PayeeName        string  `json:"payeeName"`
	PayeeIDNo        string  `json:"payeeIdNo"`
}

type WithdrawalQuoteRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type WithdrawalQuoteResponse struct {
	Amount    float64 `json:"amount"`
	FeeRate   float64 `json:"feeRate"`
	FeeAmount float64 `json:"feeAmount"`
	NetAmount float64 `json:"netAmount"`
	Currency  string  `json:"currency"`
}

type CreateWithdrawalRequest struct {
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	Password string  `json:"password" binding:"required"`
}

type ReviewWithdrawalRequest struct {
	Remark string `json:"remark"`
}

// WalletRepository defines data access methods for wallets and withdrawals.
type WalletRepository interface {
	GetWalletByUserID(userID uuid.UUID) (*Wallet, error)
	CreateWallet(wallet *Wallet) error
	CreateTransaction(tx *WalletTransaction) error
	UpdateWalletBalances(walletID uuid.UUID, rmDelta, pointsDelta float64) error
	ListTransactions(userID uuid.UUID, page, pageSize int) ([]WalletTransaction, int64, error)
	SumLifetimePointsEarned(userID uuid.UUID) (float64, error)

	// Withdrawals
	CreateWithdrawal(w *WithdrawalRequest) error
	GetWithdrawalByID(id uuid.UUID) (*WithdrawalRequest, error)
	UpdateWithdrawal(w *WithdrawalRequest) error
	ListWithdrawals(userID *uuid.UUID, status string, page, pageSize int) ([]WithdrawalRequest, int64, error)

	// Admin
	ListWallets(page, pageSize int, search string) ([]Wallet, int64, error)
	ListAllTransactions(walletID *uuid.UUID, page, pageSize int) ([]WalletTransaction, int64, error)
}

// WalletService defines business logic for wallet and withdrawal operations.
type WalletService interface {
	GetOrCreateWallet(userID uuid.UUID) (*Wallet, error)
	CreditPoints(userID uuid.UUID, amount float64, txType, refID, description string) (*WalletTransaction, error)
	CreditRM(userID uuid.UUID, amount float64, txType, refID, description string) (*WalletTransaction, error)
	DebitPoints(userID uuid.UUID, amount float64, txType, refID, description string) (*WalletTransaction, error)
	DebitRM(userID uuid.UUID, amount float64, txType, refID, description string) (*WalletTransaction, error)
	ListTransactions(userID uuid.UUID, page, pageSize int) ([]WalletTransaction, int64, error)
	SumLifetimePointsEarned(userID uuid.UUID) (float64, error)

	// Withdrawals (Redeem Rewards)
	GetWithdrawalInfo(userID uuid.UUID) (*WithdrawalInfoResponse, error)
	QuoteWithdrawal(amount float64) (*WithdrawalQuoteResponse, error)
	CreateWithdrawal(userID uuid.UUID, req CreateWithdrawalRequest) (*WithdrawalRequest, error)
	GetWithdrawal(id uuid.UUID, userID uuid.UUID, isAdmin bool) (*WithdrawalRequest, error)
	ListMyWithdrawals(userID uuid.UUID, status string, page, pageSize int) ([]WithdrawalRequest, int64, error)
	ListAllWithdrawals(status string, page, pageSize int) ([]WithdrawalRequest, int64, error)
	ApproveWithdrawal(id uuid.UUID, adminID uuid.UUID, remark string) (*WithdrawalRequest, error)
	RejectWithdrawal(id uuid.UUID, adminID uuid.UUID, remark string) (*WithdrawalRequest, error)

	// Admin
	ListWallets(page, pageSize int, search string) ([]Wallet, int64, error)
	ListAllTransactions(walletID *uuid.UUID, page, pageSize int) ([]WalletTransaction, int64, error)
}
