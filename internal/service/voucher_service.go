package service

import (
	"errors"
	"strings"
	"time"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
)

type voucherService struct {
	voucherRepo domain.VoucherRepository
	orderRepo   domain.OrderRepository
}

func NewVoucherService(voucherRepo domain.VoucherRepository, orderRepo domain.OrderRepository) domain.VoucherService {
	return &voucherService{voucherRepo: voucherRepo, orderRepo: orderRepo}
}

func (s *voucherService) CreateVoucher(adminID uuid.UUID, req domain.CreateVoucherRequest) (*domain.Voucher, error) {
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if _, err := s.voucherRepo.FindByCode(code); err == nil {
		return nil, errors.New("voucher code already exists")
	}
	if req.PerUserLimit == 0 {
		req.PerUserLimit = 1
	}
	claimable := true
	if req.IsClaimable != nil {
		claimable = *req.IsClaimable
	}
	validDays := req.ValidDaysFromClaim
	if validDays <= 0 {
		validDays = 30
	}
	v := &domain.Voucher{
		Code:               code,
		Name:               req.Name,
		Description:        req.Description,
		Type:               req.Type,
		Value:              req.Value,
		MinSpend:           req.MinSpend,
		MaxDiscount:        req.MaxDiscount,
		MaxQty:             req.MaxQty,
		PerUserLimit:       req.PerUserLimit,
		IsClaimable:        claimable,
		FirstPurchaseOnly:  req.FirstPurchaseOnly,
		ValidDaysFromClaim: validDays,
		Active:             req.Active,
		StartsAt:           req.StartsAt,
		ExpiresAt:          req.ExpiresAt,
		CreatedBy:          adminID,
	}
	return v, s.voucherRepo.Create(v)
}

func (s *voucherService) UpdateVoucher(id uuid.UUID, req domain.UpdateVoucherRequest) (*domain.Voucher, error) {
	v, err := s.voucherRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		v.Name = *req.Name
	}
	if req.Description != nil {
		v.Description = *req.Description
	}
	if req.MinSpend != nil {
		v.MinSpend = *req.MinSpend
	}
	if req.MaxDiscount != nil {
		v.MaxDiscount = *req.MaxDiscount
	}
	if req.MaxQty != nil {
		v.MaxQty = *req.MaxQty
	}
	if req.PerUserLimit != nil {
		v.PerUserLimit = *req.PerUserLimit
	}
	if req.IsClaimable != nil {
		v.IsClaimable = *req.IsClaimable
	}
	if req.FirstPurchaseOnly != nil {
		v.FirstPurchaseOnly = *req.FirstPurchaseOnly
	}
	if req.ValidDaysFromClaim != nil {
		v.ValidDaysFromClaim = *req.ValidDaysFromClaim
	}
	if req.Active != nil {
		v.Active = *req.Active
	}
	if req.StartsAt != nil {
		v.StartsAt = req.StartsAt
	}
	if req.ExpiresAt != nil {
		v.ExpiresAt = req.ExpiresAt
	}
	return v, s.voucherRepo.Update(v)
}

func (s *voucherService) DeleteVoucher(id uuid.UUID) error {
	return s.voucherRepo.Delete(id)
}

func (s *voucherService) ListVouchers(active *bool, page, pageSize int) ([]domain.Voucher, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.voucherRepo.List(active, page, pageSize)
}

func (s *voucherService) GetVoucher(id uuid.UUID) (*domain.Voucher, error) {
	return s.voucherRepo.FindByID(id)
}

func (s *voucherService) ValidateVoucher(userID uuid.UUID, req domain.ValidateVoucherRequest) (*domain.ValidateVoucherResponse, error) {
	v, err := s.voucherRepo.FindByCode(strings.ToUpper(req.Code))
	if err != nil {
		return nil, errors.New("voucher not found")
	}
	if !v.Active {
		return nil, errors.New("voucher is inactive")
	}
	now := time.Now()
	if v.StartsAt != nil && now.Before(*v.StartsAt) {
		return nil, errors.New("voucher is not yet valid")
	}
	if v.ExpiresAt != nil && now.After(*v.ExpiresAt) {
		return nil, errors.New("voucher has expired")
	}
	if v.MaxQty > 0 && v.UsedQty >= v.MaxQty {
		return nil, errors.New("voucher is fully redeemed")
	}
	if req.CartTotal < v.MinSpend {
		return nil, errors.New("minimum spend requirement not met")
	}
	if v.FirstPurchaseOnly && s.orderRepo != nil {
		_, orderCount, _ := s.orderRepo.FindByUserID(userID, 1, 1)
		if orderCount > 0 {
			return nil, errors.New("this voucher is valid for first purchase only")
		}
	}
	if v.PerUserLimit > 0 {
		used, _ := s.voucherRepo.CountRedemptionsByUser(v.ID, userID)
		if used >= v.PerUserLimit {
			return nil, errors.New("you have already used this voucher")
		}
	}
	discount := 0.0
	if v.Type == domain.VoucherTypePercent {
		discount = req.CartTotal * (v.Value / 100)
		if v.MaxDiscount > 0 && discount > v.MaxDiscount {
			discount = v.MaxDiscount
		}
	} else {
		discount = v.Value
		if discount > req.CartTotal {
			discount = req.CartTotal
		}
	}
	return &domain.ValidateVoucherResponse{
		Voucher:        *v,
		DiscountAmount: discount,
		FinalTotal:     req.CartTotal - discount,
	}, nil
}

func (s *voucherService) ApplyVoucher(userID uuid.UUID, voucherCode string, orderID uuid.UUID, originalAmount, discountAmount, finalAmount float64) error {
	v, err := s.voucherRepo.FindByCode(voucherCode)
	if err != nil {
		return err
	}
	if err := s.voucherRepo.IncrementUsed(v.ID); err != nil {
		return err
	}
	_ = s.voucherRepo.MarkUserVoucherUsed(v.ID, userID, orderID)
	return s.voucherRepo.CreateRedemption(&domain.VoucherRedemption{
		VoucherID:      v.ID,
		UserID:         userID,
		OrderID:        orderID,
		OriginalAmount: originalAmount,
		DiscountAmount: discountAmount,
		FinalAmount:    finalAmount,
	})
}

func (s *voucherService) ListRedemptions(voucherID *uuid.UUID, page, pageSize int) ([]domain.VoucherRedemption, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.voucherRepo.ListRedemptions(voucherID, page, pageSize)
}

func (s *voucherService) GetMyVouchers(userID uuid.UUID) (*domain.MyVouchersResponse, error) {
	held, err := s.voucherRepo.ListUserVouchers(userID)
	if err != nil {
		return nil, err
	}
	heldMap := make(map[uuid.UUID]bool, len(held))
	for _, h := range held {
		heldMap[h.VoucherID] = true
	}

	allClaimable, err := s.voucherRepo.ListClaimableVouchers()
	if err != nil {
		return nil, err
	}
	claimable := make([]domain.Voucher, 0)
	for _, v := range allClaimable {
		if heldMap[v.ID] {
			continue
		}
		if v.MaxQty > 0 && v.UsedQty >= v.MaxQty {
			continue
		}
		if v.PerUserLimit > 0 {
			used, _ := s.voucherRepo.CountRedemptionsByUser(v.ID, userID)
			if used >= v.PerUserLimit {
				continue
			}
		}
		claimable = append(claimable, v)
	}
	return &domain.MyVouchersResponse{
		Held:      held,
		Claimable: claimable,
	}, nil
}

func (s *voucherService) ClaimVoucher(userID uuid.UUID, code string) (*domain.UserVoucher, error) {
	v, err := s.voucherRepo.FindByCode(strings.ToUpper(strings.TrimSpace(code)))
	if err != nil {
		return nil, errors.New("voucher not found")
	}
	if !v.Active || !v.IsClaimable {
		return nil, errors.New("voucher is not available to claim")
	}
	if v.MaxQty > 0 && v.UsedQty >= v.MaxQty {
		return nil, errors.New("voucher has been fully claimed")
	}
	if _, err := s.voucherRepo.FindUnusedUserVoucher(v.ID, userID); err == nil {
		return nil, errors.New("you have already claimed this voucher")
	}
	if v.PerUserLimit > 0 {
		used, _ := s.voucherRepo.CountRedemptionsByUser(v.ID, userID)
		if used >= v.PerUserLimit {
			return nil, errors.New("you have reached the limit for this voucher")
		}
	}

	var exp *time.Time
	if v.ExpiresAt != nil {
		exp = v.ExpiresAt
	} else if v.ValidDaysFromClaim > 0 {
		t := time.Now().AddDate(0, 0, v.ValidDaysFromClaim)
		exp = &t
	}

	uv := &domain.UserVoucher{
		VoucherID: v.ID,
		UserID:    userID,
		ExpiresAt: exp,
	}
	if err := s.voucherRepo.CreateUserVoucher(uv); err != nil {
		return nil, err
	}
	uv.Voucher = v
	return uv, nil
}
