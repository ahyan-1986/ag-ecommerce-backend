package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole defines access level in the Ascend Group platform.
type UserRole string

const (
	RoleSuperAdmin UserRole = "super_admin"
	RoleAdmin      UserRole = "admin"
	RoleDealer     UserRole = "dealer"
	RoleCustomer   UserRole = "customer"
)

// MemberRank defines the 4-tier Aameba MLM rank ladder: M -> S1 -> S2 -> S3.
type MemberRank string

const (
	RankMember    MemberRank = "M"
	RankOneStar   MemberRank = "S1"
	RankTwoStar   MemberRank = "S2"
	RankThreeStar MemberRank = "S3"
)

func (r MemberRank) Ordinal() int {
	switch r {
	case RankMember:
		return 1
	case RankOneStar:
		return 2
	case RankTwoStar:
		return 3
	case RankThreeStar:
		return 4
	default:
		return 0
	}
}

func (r MemberRank) Label() string {
	switch r {
	case RankMember:
		return "Member"
	case RankOneStar:
		return "One Star"
	case RankTwoStar:
		return "Two Star"
	case RankThreeStar:
		return "Three Star"
	default:
		return "Member"
	}
}

// User is the central identity entity used for all portal logins.
type User struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email         string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash  string     `gorm:"type:varchar(255);not null" json:"-"`
	FirstName     string     `gorm:"type:varchar(100);not null;default:''" json:"firstName"`
	LastName      string     `gorm:"type:varchar(100);not null;default:''" json:"lastName"`
	Phone         string     `gorm:"type:varchar(30);uniqueIndex;default:''" json:"phone"`
	AvatarURL     string     `gorm:"type:varchar(500);default:''" json:"avatarUrl"`
	DateOfBirth   *time.Time `gorm:"type:date" json:"dateOfBirth,omitempty"`
	Gender        string     `gorm:"type:varchar(20);default:''" json:"gender"`
	Role          UserRole   `gorm:"type:varchar(30);not null;default:'customer'" json:"role"`
	IsActive      bool       `gorm:"not null;default:true" json:"isActive"`
	EmailVerified bool       `gorm:"not null;default:false" json:"emailVerified"`
	// ReferralCode is this user's own shareable code (auto-generated on register)
	ReferralCode string `gorm:"type:varchar(20);uniqueIndex;not null;default:''" json:"referralCode"`
	// MemberCode is the human-readable member ID (e.g. AG10001)
	MemberCode string `gorm:"type:varchar(30);index;default:''" json:"memberCode"`
	// MLM Rank Ladder (M, S1, S2, S3)
	MemberRank  MemberRank `gorm:"type:varchar(10);not null;default:'M'" json:"memberRank"`
	HighestRank MemberRank `gorm:"type:varchar(10);not null;default:'M'" json:"highestRank"`
	// ReferredByCode is the code entered by user at registration (optional)
	ReferredByCode string     `gorm:"type:varchar(20);default:''" json:"referredByCode"`
	ReferredByID   *uuid.UUID `gorm:"type:uuid" json:"referredById,omitempty"`

	// Bank / Payee details (required for Redeem Rewards / Withdrawals)
	BankName      string `gorm:"type:varchar(120);default:''" json:"bankName"`
	BankAccountNo string `gorm:"type:varchar(60);default:''" json:"bankAccountNo"`
	PayeeName     string `gorm:"type:varchar(150);default:''" json:"payeeName"`
	PayeeIDNo     string `gorm:"type:varchar(60);default:''" json:"payeeIdNo"`

	// Dealer-specific
	CompanyName    string `gorm:"type:varchar(255);default:''" json:"companyName"`
	SSMNumber      string `gorm:"type:varchar(100);default:''" json:"ssmNumber"`
	DealerApproved bool   `gorm:"not null;default:false" json:"dealerApproved"`

	// Address
	AddressLine1 string `gorm:"type:varchar(255);default:''" json:"addressLine1"`
	AddressLine2 string `gorm:"type:varchar(255);default:''" json:"addressLine2"`
	City         string `gorm:"type:varchar(100);default:''" json:"city"`
	State        string `gorm:"type:varchar(100);default:''" json:"state"`
	Postcode     string `gorm:"type:varchar(20);default:''" json:"postcode"`
	Country      string `gorm:"type:varchar(100);not null;default:'Malaysia'" json:"country"`

	// Email verification token
	EmailVerifyToken    string     `gorm:"type:varchar(100);default:''" json:"-"`
	EmailVerifyExpiry   *time.Time `gorm:"type:timestamptz" json:"-"`
	PasswordResetToken  string     `gorm:"type:varchar(100);default:''" json:"-"`
	PasswordResetExpiry *time.Time `gorm:"type:timestamptz" json:"-"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string     { return "users" }
func (u *User) IsSuperAdmin() bool { return u.Role == RoleSuperAdmin }
func (u *User) IsAdmin() bool      { return u.Role == RoleAdmin || u.Role == RoleSuperAdmin }
func (u *User) IsDealer() bool     { return u.Role == RoleDealer }
func (u *User) IsCustomer() bool   { return u.Role == RoleCustomer }

type UserPublicProfile struct {
	ID             uuid.UUID `json:"id"`
	Email          string    `json:"email"`
	FirstName      string    `json:"firstName"`
	LastName       string    `json:"lastName"`
	Phone          string    `json:"phone"`
	AvatarURL      string    `json:"avatarUrl"`
	Role           UserRole  `json:"role"`
	IsActive       bool      `json:"isActive"`
	EmailVerified  bool      `json:"emailVerified"`
	ReferralCode   string     `json:"referralCode"`
	MemberCode     string     `json:"memberCode"`
	MemberRank     MemberRank `json:"memberRank"`
	HighestRank    MemberRank `json:"highestRank"`
	ReferredByCode string     `json:"referredByCode"`
	BankName       string     `json:"bankName"`
	BankAccountNo  string     `json:"bankAccountNo"`
	PayeeName      string     `json:"payeeName"`
	PayeeIDNo      string     `json:"payeeIdNo"`
	CompanyName    string     `json:"companyName"`
	DealerApproved bool      `json:"dealerApproved"`
	AddressLine1   string    `json:"addressLine1"`
	AddressLine2   string    `json:"addressLine2"`
	City           string    `json:"city"`
	State          string    `json:"state"`
	Postcode       string    `json:"postcode"`
	Country        string    `json:"country"`
	DateOfBirth    *time.Time `json:"dateOfBirth,omitempty"`
	Gender         string    `json:"gender"`
	CreatedAt      time.Time `json:"createdAt"`
}

// CRMUserStats contains order + spend aggregates for the CRM 360 view.
type CRMUserStats struct {
	UserID           string  `json:"userId"`
	TotalOrders      int64   `json:"totalOrders"`
	TotalSpend       float64 `json:"totalSpend"`
	LastOrderAt      string  `json:"lastOrderAt"`
	DownlineCount    int64   `json:"downlineCount"`
	CommissionEarned float64 `json:"commissionEarned"`
}

// CRMSummary is the CRM overview KPI block.
type CRMSummary struct {
	TotalCustomers   int64   `json:"totalCustomers"`
	ActiveCustomers  int64   `json:"activeCustomers"`
	NewThisMonth     int64   `json:"newThisMonth"`
	TotalRevenue     float64 `json:"totalRevenue"`
	AvgOrderValue    float64 `json:"avgOrderValue"`
	WithReferralCode int64   `json:"withReferralCode"`
}

// ── Auth DTOs ─────────────────────────────────────────────────────────────────

type RegisterRequest struct {
	Email        string   `json:"email" binding:"required,email"`
	Password     string   `json:"password" binding:"required,min=8"`
	FirstName    string   `json:"firstName" binding:"required"`
	LastName     string   `json:"lastName" binding:"required"`
	Phone        string   `json:"phone" binding:"required"`
	Role         UserRole `json:"role"`         // customer (default) or dealer
	ReferralCode string   `json:"referralCode"` // optional — referrer's code
	CompanyName  string   `json:"companyName"`  // required if dealer
	SSMNumber    string   `json:"ssmNumber"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string            `json:"accessToken"`
	RefreshToken string            `json:"refreshToken"`
	User         UserPublicProfile `json:"user"`
}

type UpdateProfileRequest struct {
	FirstName    *string    `json:"firstName"`
	LastName     *string    `json:"lastName"`
	Phone        *string    `json:"phone"`
	DateOfBirth  *time.Time `json:"dateOfBirth"`
	Gender       *string    `json:"gender"`
	AddressLine1 *string    `json:"addressLine1"`
	AddressLine2 *string    `json:"addressLine2"`
	City         *string    `json:"city"`
	State        *string    `json:"state"`
	Postcode      *string    `json:"postcode"`
	CompanyName   *string    `json:"companyName"`
	BankName      *string    `json:"bankName"`
	BankAccountNo *string    `json:"bankAccountNo"`
	PayeeName     *string    `json:"payeeName"`
	PayeeIDNo     *string    `json:"payeeIdNo"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=8"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// ── User Service & Repository interfaces ─────────────────────────────────────

type UserRepository interface {
	Create(user *User) error
	FindByID(id uuid.UUID) (*User, error)
	FindByEmail(email string) (*User, error)
	FindByPhone(phone string) (*User, error)
	FindByReferralCode(code string) (*User, error)
	FindByEmailVerifyToken(token string) (*User, error)
	FindByPasswordResetToken(token string) (*User, error)
	Update(user *User) error
	Delete(id uuid.UUID) error
	List(role UserRole, search string, page, pageSize int) ([]User, int64, error)
	Count() (int64, error)
	CountByRole(role UserRole) (int64, error)
	GetCRMSummary() (*CRMSummary, error)
	GetUserCRMStats(userID uuid.UUID) (*CRMUserStats, error)
}

type UserService interface {
	GetProfile(userID uuid.UUID) (*UserPublicProfile, error)
	UpdateProfile(userID uuid.UUID, req UpdateProfileRequest) (*UserPublicProfile, error)
	ChangePassword(userID uuid.UUID, req ChangePasswordRequest) error
	UploadAvatar(userID uuid.UUID, data []byte, contentType string) (string, error)
	ListUsers(role UserRole, search string, page, pageSize int) ([]UserPublicProfile, int64, error)
	GetUser(id uuid.UUID) (*UserPublicProfile, error)
	SetActive(id uuid.UUID, active bool) error
	ApproveDealerApplication(id uuid.UUID) error
	GetCRMSummary() (*CRMSummary, error)
	GetUserCRMStats(userID uuid.UUID) (*CRMUserStats, error)
}
