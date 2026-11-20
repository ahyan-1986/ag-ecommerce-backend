package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/config"
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type authService struct {
	userRepo     domain.UserRepository
	emailSvc     domain.EmailService
	affiliateSvc domain.AffiliateService
	cfg          *config.Config
}

func NewAuthService(userRepo domain.UserRepository, emailSvc domain.EmailService, affiliateSvc domain.AffiliateService, cfg *config.Config) *authService {
	return &authService{userRepo: userRepo, emailSvc: emailSvc, affiliateSvc: affiliateSvc, cfg: cfg}
}

func (s *authService) Register(req domain.RegisterRequest) (*domain.AuthResponse, error) {
	// Validate role
	role := req.Role
	if role == "" {
		role = domain.RoleCustomer
	}
	if role != domain.RoleCustomer && role != domain.RoleDealer {
		return nil, errors.New("invalid role: must be customer or dealer")
	}

	// Check email uniqueness
	if _, err := s.userRepo.FindByEmail(req.Email); err == nil {
		return nil, errors.New("email already registered")
	}

	// Check phone uniqueness
	if req.Phone != "" {
		if _, err := s.userRepo.FindByPhone(req.Phone); err == nil {
			return nil, errors.New("phone number already registered")
		}
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Resolve referrer
	var referredByID *uuid.UUID
	if req.ReferralCode != "" {
		referrer, err := s.userRepo.FindByReferralCode(req.ReferralCode)
		if err == nil {
			referredByID = &referrer.ID
		}
	}

	// Generate email verify token
	verifyToken, _ := generateSecureToken(32)
	verifyExpiry := time.Now().Add(48 * time.Hour)

	refCode := generateReferralCode()
	user := &domain.User{
		Email:             req.Email,
		PasswordHash:      string(hash),
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		Phone:             req.Phone,
		Role:              role,
		ReferralCode:      refCode,
		MemberCode:        "AG" + refCode[:6],
		MemberRank:        domain.RankMember,
		HighestRank:       domain.RankMember,
		ReferredByCode:    req.ReferralCode,
		ReferredByID:      referredByID,
		CompanyName:       req.CompanyName,
		SSMNumber:         req.SSMNumber,
		EmailVerifyToken:  verifyToken,
		EmailVerifyExpiry: &verifyExpiry,
		IsActive:          true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// Release any held guest-order loyalty points for this email (Aameba GuestRebate)
	if s.affiliateSvc != nil {
		go func(u domain.User) {
			_ = s.affiliateSvc.ReleaseGuestHeldPoints(&u)
		}(*user)
	}

	// Send verification email (non-blocking)
	if s.emailSvc != nil {
		verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.cfg.Email.AppURL, verifyToken)
		go s.emailSvc.SendEmailVerification(user.Email, user.FirstName, verifyURL)
	}

	// Issue tokens
	return s.issueTokens(user)
}

func (s *authService) Login(req domain.LoginRequest) (*domain.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	return s.issueTokens(user)
}

func (s *authService) VerifyEmail(token string) error {
	user, err := s.userRepo.FindByEmailVerifyToken(token)
	if err != nil {
		return errors.New("invalid or expired verification link")
	}
	if user.EmailVerifyExpiry != nil && user.EmailVerifyExpiry.Before(time.Now()) {
		return errors.New("verification link has expired")
	}
	user.EmailVerified = true
	user.EmailVerifyToken = ""
	user.EmailVerifyExpiry = nil
	return s.userRepo.Update(user)
}

func (s *authService) ForgotPassword(req domain.ForgotPasswordRequest) error {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		// Silently succeed to prevent email enumeration
		return nil
	}
	token, _ := generateSecureToken(32)
	expiry := time.Now().Add(1 * time.Hour)
	user.PasswordResetToken = token
	user.PasswordResetExpiry = &expiry
	if err := s.userRepo.Update(user); err != nil {
		return err
	}
	if s.emailSvc != nil {
		resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.cfg.Email.AppURL, token)
		go s.emailSvc.SendPasswordReset(user.Email, user.FirstName, resetURL)
	}
	return nil
}

func (s *authService) ResetPassword(req domain.ResetPasswordRequest) error {
	user, err := s.userRepo.FindByPasswordResetToken(req.Token)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}
	if user.PasswordResetExpiry != nil && user.PasswordResetExpiry.Before(time.Now()) {
		return errors.New("reset token has expired")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hash)
	user.PasswordResetToken = ""
	user.PasswordResetExpiry = nil
	return s.userRepo.Update(user)
}

func (s *authService) RefreshToken(refreshToken string) (*domain.AuthResponse, error) {
	userID, err := s.parseRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}
	user, err := s.userRepo.FindByID(userID)
	if err != nil || !user.IsActive {
		return nil, errors.New("user not found")
	}
	return s.issueTokens(user)
}

// ── Token helpers ─────────────────────────────────────────────────────────────

type ascendJWTClaims struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
	Type   string `json:"type"` // access | refresh
	jwt.RegisteredClaims
}

func (s *authService) issueTokens(user *domain.User) (*domain.AuthResponse, error) {
	accessToken, err := s.signToken(user, "access", s.cfg.JWT.GetTokenExpiry())
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.signToken(user, "refresh", s.cfg.JWT.GetRefreshExpiry())
	if err != nil {
		return nil, err
	}

	profile := toPublicProfile(user)
	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         profile,
	}, nil
}

func (s *authService) signToken(user *domain.User, tokenType string, expiry time.Duration) (string, error) {
	claims := ascendJWTClaims{
		UserID: user.ID.String(),
		Role:   string(user.Role),
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ascendgroup",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}

func (s *authService) parseRefreshToken(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &ascendJWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWT.Secret), nil
	})
	if err != nil || !token.Valid {
		return uuid.UUID{}, errors.New("invalid token")
	}
	claims, ok := token.Claims.(*ascendJWTClaims)
	if !ok || claims.Type != "refresh" {
		return uuid.UUID{}, errors.New("not a refresh token")
	}
	return uuid.Parse(claims.UserID)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func generateSecureToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func generateReferralCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return strings.ToUpper(hex.EncodeToString(b))
}

func toPublicProfile(u *domain.User) domain.UserPublicProfile {
	rank := u.HighestRank
	if rank == "" {
		rank = u.MemberRank
	}
	if rank == "" {
		rank = domain.RankMember
	}
	return domain.UserPublicProfile{
		ID:             u.ID,
		Email:          u.Email,
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		Phone:          u.Phone,
		AvatarURL:      u.AvatarURL,
		Role:           u.Role,
		IsActive:       u.IsActive,
		EmailVerified:  u.EmailVerified,
		ReferralCode:   u.ReferralCode,
		MemberCode:     u.MemberCode,
		MemberRank:     rank,
		HighestRank:    rank,
		ReferredByCode: u.ReferredByCode,
		BankName:       u.BankName,
		BankAccountNo:  u.BankAccountNo,
		PayeeName:      u.PayeeName,
		PayeeIDNo:      u.PayeeIDNo,
		CompanyName:    u.CompanyName,
		DealerApproved: u.DealerApproved,
		AddressLine1:   u.AddressLine1,
		AddressLine2:   u.AddressLine2,
		City:           u.City,
		State:          u.State,
		Postcode:       u.Postcode,
		Country:        u.Country,
		DateOfBirth:    u.DateOfBirth,
		Gender:         u.Gender,
		CreatedAt:      u.CreatedAt,
	}
}
