package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	VoucherTypePercent = "PERCENT"
	VoucherTypeCash    = "CASH"
)

// Voucher is a discount code usable at checkout or claimable in My Vouchers.
type Voucher struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code               string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Name               string     `gorm:"type:varchar(255);not null" json:"name"`
	Description        string     `gorm:"type:text;default:''" json:"description"`
	Type               string     `gorm:"type:varchar(10);not null" json:"type"`    // PERCENT | CASH
	Value              float64    `gorm:"type:numeric(10,2);not null" json:"value"` // % or MYR amount
	MinSpend           float64    `gorm:"type:numeric(10,2);default:0" json:"minSpend"`
	MaxDiscount        float64    `gorm:"type:numeric(10,2);default:0" json:"maxDiscount"` // 0 = unlimited
	MaxQty             int        `gorm:"default:0" json:"maxQty"`                         // 0 = unlimited
	UsedQty            int        `gorm:"default:0" json:"usedQty"`
	PerUserLimit       int        `gorm:"default:1" json:"perUserLimit"`
	IsClaimable        bool       `gorm:"not null;default:true" json:"isClaimable"`
	FirstPurchaseOnly  bool       `gorm:"not null;default:false" json:"firstPurchaseOnly"`
	ValidDaysFromClaim int        `gorm:"not null;default:30" json:"validDaysFromClaim"`
	Active             bool       `gorm:"not null;default:true" json:"active"`
	StartsAt           *time.Time `gorm:"type:timestamptz" json:"startsAt,omitempty"`
	ExpiresAt          *time.Time `gorm:"type:timestamptz" json:"expiresAt,omitempty"`
	CreatedBy          uuid.UUID  `gorm:"type:uuid;not null" json:"createdBy"`
	CreatedAt          time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Voucher) TableName() string { return "vouchers" }

// UserVoucher represents a voucher claimed into a member's wallet (Aameba My Vouchers "Ready to use").
type UserVoucher struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	VoucherID uuid.UUID  `gorm:"type:uuid;not null;index" json:"voucherId"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"userId"`
	ClaimedAt time.Time  `gorm:"autoCreateTime" json:"claimedAt"`
	ExpiresAt *time.Time `gorm:"type:timestamptz" json:"expiresAt,omitempty"`
	UsedAt    *time.Time `gorm:"type:timestamptz" json:"usedAt,omitempty"`
	OrderID   *uuid.UUID `gorm:"type:uuid" json:"orderId,omitempty"`

	Voucher *Voucher `gorm:"foreignKey:VoucherID" json:"voucher,omitempty"`
}

func (UserVoucher) TableName() string { return "user_vouchers" }

// VoucherRedemption records each use of a voucher.
type VoucherRedemption struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	VoucherID      uuid.UUID `gorm:"type:uuid;not null;index" json:"voucherId"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	OrderID        uuid.UUID `gorm:"type:uuid;not null" json:"orderId"`
	OriginalAmount float64   `gorm:"type:numeric(12,2)" json:"originalAmount"`
	DiscountAmount float64   `gorm:"type:numeric(12,2)" json:"discountAmount"`
	FinalAmount    float64   `gorm:"type:numeric(12,2)" json:"finalAmount"`
	RedeemedAt     time.Time `gorm:"autoCreateTime" json:"redeemedAt"`

	Voucher *Voucher `gorm:"foreignKey:VoucherID" json:"voucher,omitempty"`
}

func (VoucherRedemption) TableName() string { return "voucher_redemptions" }

// ── Request / Response DTOs ───────────────────────────────────────────────────

type CreateVoucherRequest struct {
	Code               string     `json:"code" binding:"required"`
	Name               string     `json:"name" binding:"required"`
	Description        string     `json:"description"`
	Type               string     `json:"type" binding:"required,oneof=PERCENT CASH"`
	Value              float64    `json:"value" binding:"required,gt=0"`
	MinSpend           float64    `json:"minSpend"`
	MaxDiscount        float64    `json:"maxDiscount"`
	MaxQty             int        `json:"maxQty"`
	PerUserLimit       int        `json:"perUserLimit"`
	IsClaimable        *bool      `json:"isClaimable"`
	FirstPurchaseOnly  bool       `json:"firstPurchaseOnly"`
	ValidDaysFromClaim int        `json:"validDaysFromClaim"`
	Active             bool       `json:"active"`
	StartsAt           *time.Time `json:"startsAt"`
	ExpiresAt          *time.Time `json:"expiresAt"`
}

type UpdateVoucherRequest struct {
	Name               *string    `json:"name"`
	Description        *string    `json:"description"`
	MinSpend           *float64   `json:"minSpend"`
	MaxDiscount        *float64   `json:"maxDiscount"`
	MaxQty             *int       `json:"maxQty"`
	PerUserLimit       *int       `json:"perUserLimit"`
	IsClaimable        *bool      `json:"isClaimable"`
	FirstPurchaseOnly  *bool      `json:"firstPurchaseOnly"`
	ValidDaysFromClaim *int       `json:"validDaysFromClaim"`
	Active             *bool      `json:"active"`
	StartsAt           *time.Time `json:"startsAt"`
	ExpiresAt          *time.Time `json:"expiresAt"`
}

type ValidateVoucherRequest struct {
	Code      string  `json:"code" binding:"required"`
	CartTotal float64 `json:"cartTotal" binding:"required,gt=0"`
}

type ValidateVoucherResponse struct {
	Voucher        Voucher `json:"voucher"`
	DiscountAmount float64 `json:"discountAmount"`
	FinalTotal     float64 `json:"finalTotal"`
}

type MyVouchersResponse struct {
	Held      []UserVoucher `json:"held"`
	Claimable []Voucher     `json:"claimable"`
}

// ── Interfaces ────────────────────────────────────────────────────────────────

type VoucherRepository interface {
	Create(v *Voucher) error
	FindByID(id uuid.UUID) (*Voucher, error)
	FindByCode(code string) (*Voucher, error)
	Update(v *Voucher) error
	Delete(id uuid.UUID) error
	List(active *bool, page, pageSize int) ([]Voucher, int64, error)
	IncrementUsed(id uuid.UUID) error
	CreateRedemption(r *VoucherRedemption) error
	CountRedemptionsByUser(voucherID, userID uuid.UUID) (int, error)
	ListRedemptions(voucherID *uuid.UUID, page, pageSize int) ([]VoucherRedemption, int64, error)
	// Member wallet vouchers (Aameba My Vouchers)
	CreateUserVoucher(uv *UserVoucher) error
	ListUserVouchers(userID uuid.UUID) ([]UserVoucher, error)
	FindUnusedUserVoucher(voucherID, userID uuid.UUID) (*UserVoucher, error)
	MarkUserVoucherUsed(voucherID, userID, orderID uuid.UUID) error
	ListClaimableVouchers() ([]Voucher, error)
}

type VoucherService interface {
	CreateVoucher(adminID uuid.UUID, req CreateVoucherRequest) (*Voucher, error)
	UpdateVoucher(id uuid.UUID, req UpdateVoucherRequest) (*Voucher, error)
	DeleteVoucher(id uuid.UUID) error
	ListVouchers(active *bool, page, pageSize int) ([]Voucher, int64, error)
	GetVoucher(id uuid.UUID) (*Voucher, error)
	ValidateVoucher(userID uuid.UUID, req ValidateVoucherRequest) (*ValidateVoucherResponse, error)
	ApplyVoucher(userID uuid.UUID, voucherCode string, orderID uuid.UUID, originalAmount, discountAmount, finalAmount float64) error
	ListRedemptions(voucherID *uuid.UUID, page, pageSize int) ([]VoucherRedemption, int64, error)
	// Member My Vouchers
	GetMyVouchers(userID uuid.UUID) (*MyVouchersResponse, error)
	ClaimVoucher(userID uuid.UUID, code string) (*UserVoucher, error)
}
