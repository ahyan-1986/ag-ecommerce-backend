package service

import (
	"encoding/json"
	"errors"
	"math"
	"strings"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
)

type shippingService struct {
	repo domain.ShippingRepository
}

func NewShippingService(repo domain.ShippingRepository) domain.ShippingService {
	return &shippingService{repo: repo}
}

// ── Zones ─────────────────────────────────────────────────────────────────────

func (s *shippingService) ListZones() ([]domain.ShippingZone, error) {
	return s.repo.ListZones()
}

func (s *shippingService) CreateZone(req domain.CreateShippingZoneRequest) (*domain.ShippingZone, error) {
	countries, _ := json.Marshal(req.Countries)
	states, _ := json.Marshal(req.States)
	z := &domain.ShippingZone{
		Name:        req.Name,
		Description: req.Description,
		Countries:   string(countries),
		States:      string(states),
		IsActive:    req.IsActive,
		SortOrder:   req.SortOrder,
	}
	if err := s.repo.CreateZone(z); err != nil {
		return nil, err
	}
	return z, nil
}

func (s *shippingService) UpdateZone(id uuid.UUID, req domain.UpdateShippingZoneRequest) (*domain.ShippingZone, error) {
	z, err := s.repo.GetZone(id)
	if err != nil {
		return nil, errors.New("zone not found")
	}
	if req.Name != nil {
		z.Name = *req.Name
	}
	if req.Description != nil {
		z.Description = *req.Description
	}
	if req.Countries != nil {
		b, _ := json.Marshal(req.Countries)
		z.Countries = string(b)
	}
	if req.States != nil {
		b, _ := json.Marshal(req.States)
		z.States = string(b)
	}
	if req.IsActive != nil {
		z.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		z.SortOrder = *req.SortOrder
	}
	if err := s.repo.UpdateZone(z); err != nil {
		return nil, err
	}
	return z, nil
}

func (s *shippingService) DeleteZone(id uuid.UUID) error {
	return s.repo.DeleteZone(id)
}

// ── Rates ─────────────────────────────────────────────────────────────────────

func (s *shippingService) ListRates(zoneID *uuid.UUID) ([]domain.ShippingRate, error) {
	return s.repo.ListRates(zoneID)
}

func (s *shippingService) CreateRate(req domain.CreateShippingRateRequest) (*domain.ShippingRate, error) {
	weightTiers, _ := json.Marshal(req.WeightTiers)
	priceTiers, _ := json.Marshal(req.PriceTiers)

	divisor := req.VolumetricDivisor
	if divisor == 0 {
		divisor = 5000 // industry standard default
	}

	rate := &domain.ShippingRate{
		ZoneID:    req.ZoneID,
		Name:      req.Name,
		Carrier:   req.Carrier,
		RateType:  req.RateType,
		IsActive:  req.IsActive,
		SortOrder: req.SortOrder,

		FlatRateRM: req.FlatRateRM,

		WeightTiers: string(weightTiers),
		PerKgRM:     req.PerKgRM,
		FreeKgLimit: req.FreeKgLimit,

		VolumetricDivisor: divisor,

		PriceTiers: string(priceTiers),

		PerItemRM:  req.PerItemRM,
		MaxItemFee: req.MaxItemFee,

		CODFeeRM:      req.CODFeeRM,
		CODFeePercent: req.CODFeePercent,
		CODMaxFeeRM:   req.CODMaxFeeRM,

		MinOrderRM: req.MinOrderRM,
		MaxOrderRM: req.MaxOrderRM,

		EstDaysMin: req.EstDaysMin,
		EstDaysMax: req.EstDaysMax,
		CutoffTime: req.CutoffTime,
	}
	if err := s.repo.CreateRate(rate); err != nil {
		return nil, err
	}
	return rate, nil
}

func (s *shippingService) UpdateRate(id uuid.UUID, req domain.UpdateShippingRateRequest) (*domain.ShippingRate, error) {
	rate, err := s.repo.GetRate(id)
	if err != nil {
		return nil, errors.New("rate not found")
	}
	if req.Name != nil {
		rate.Name = *req.Name
	}
	if req.Carrier != nil {
		rate.Carrier = *req.Carrier
	}
	if req.RateType != nil {
		rate.RateType = *req.RateType
	}
	if req.IsActive != nil {
		rate.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		rate.SortOrder = *req.SortOrder
	}
	if req.FlatRateRM != nil {
		rate.FlatRateRM = *req.FlatRateRM
	}
	if req.WeightTiers != nil {
		b, _ := json.Marshal(req.WeightTiers)
		rate.WeightTiers = string(b)
	}
	if req.PerKgRM != nil {
		rate.PerKgRM = *req.PerKgRM
	}
	if req.FreeKgLimit != nil {
		rate.FreeKgLimit = *req.FreeKgLimit
	}
	if req.VolumetricDivisor != nil {
		rate.VolumetricDivisor = *req.VolumetricDivisor
	}
	if req.PriceTiers != nil {
		b, _ := json.Marshal(req.PriceTiers)
		rate.PriceTiers = string(b)
	}
	if req.PerItemRM != nil {
		rate.PerItemRM = *req.PerItemRM
	}
	if req.MaxItemFee != nil {
		rate.MaxItemFee = *req.MaxItemFee
	}
	if req.CODFeeRM != nil {
		rate.CODFeeRM = *req.CODFeeRM
	}
	if req.CODFeePercent != nil {
		rate.CODFeePercent = *req.CODFeePercent
	}
	if req.CODMaxFeeRM != nil {
		rate.CODMaxFeeRM = *req.CODMaxFeeRM
	}
	if req.MinOrderRM != nil {
		rate.MinOrderRM = *req.MinOrderRM
	}
	if req.MaxOrderRM != nil {
		rate.MaxOrderRM = *req.MaxOrderRM
	}
	if req.EstDaysMin != nil {
		rate.EstDaysMin = *req.EstDaysMin
	}
	if req.EstDaysMax != nil {
		rate.EstDaysMax = *req.EstDaysMax
	}
	if req.CutoffTime != nil {
		rate.CutoffTime = *req.CutoffTime
	}
	if err := s.repo.UpdateRate(rate); err != nil {
		return nil, err
	}
	return rate, nil
}

func (s *shippingService) DeleteRate(id uuid.UUID) error {
	return s.repo.DeleteRate(id)
}

// ── Calculate ─────────────────────────────────────────────────────────────────

func (s *shippingService) CalculateRates(req domain.CalculateShippingRequest) ([]domain.CalculatedRate, error) {
	zones, err := s.repo.ListZones()
	if err != nil {
		return nil, err
	}

	country := strings.ToUpper(strings.TrimSpace(req.Country))
	stateLower := strings.ToLower(strings.TrimSpace(req.State))

	var results []domain.CalculatedRate

	for _, zone := range zones {
		if !zone.IsActive {
			continue
		}
		if !zoneMatches(zone, country, stateLower) {
			continue
		}
		for _, rate := range zone.Rates {
			if !rate.IsActive {
				continue
			}
			// Check order amount conditions
			if rate.MinOrderRM > 0 && req.OrderRM < rate.MinOrderRM {
				continue
			}
			if rate.MaxOrderRM > 0 && req.OrderRM > rate.MaxOrderRM {
				continue
			}
			cost := CalcCost(rate, req)
			results = append(results, domain.CalculatedRate{
				RateID:     rate.ID,
				ZoneName:   zone.Name,
				Name:       rate.Name,
				Carrier:    rate.Carrier,
				RateType:   string(rate.RateType),
				CostRM:     cost,
				EstDaysMin: rate.EstDaysMin,
				EstDaysMax: rate.EstDaysMax,
				CutoffTime: rate.CutoffTime,
			})
		}
	}
	return results, nil
}

// ── Helper: zone matching ─────────────────────────────────────────────────────

func zoneMatches(zone domain.ShippingZone, country, stateLower string) bool {
	var countries []string
	json.Unmarshal([]byte(zone.Countries), &countries) //nolint:errcheck
	if len(countries) > 0 {
		found := false
		for _, c := range countries {
			if strings.EqualFold(c, country) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	var states []string
	json.Unmarshal([]byte(zone.States), &states) //nolint:errcheck
	if len(states) > 0 && stateLower != "" {
		found := false
		for _, st := range states {
			if strings.EqualFold(st, stateLower) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func (s *shippingService) GetRate(id uuid.UUID) (*domain.ShippingRate, error) {
	return s.repo.GetRate(id)
}

// ── Helper: cost calculation per rate type ────────────────────────────────────

func CalcCost(rate domain.ShippingRate, req domain.CalculateShippingRequest) float64 {
	switch rate.RateType {

	case domain.RateTypeFree:
		// Free if minOrder met, else charge flatRateRM as fallback
		if rate.MinOrderRM > 0 && req.OrderRM < rate.MinOrderRM {
			return rate.FlatRateRM
		}
		return 0

	case domain.RateTypeByWeight:
		// Tiered by actual weight — pick the matching tier
		var tiers []domain.WeightTier
		json.Unmarshal([]byte(rate.WeightTiers), &tiers) //nolint:errcheck
		for _, t := range tiers {
			if req.WeightKg >= t.MinKg && (t.MaxKg == 0 || req.WeightKg <= t.MaxKg) {
				return t.RateRM
			}
		}
		return rate.FlatRateRM // fallback

	case domain.RateTypeFlatPlusKg:
		// Base fee + per-kg above the free threshold
		surchargeKg := math.Max(0, req.WeightKg-rate.FreeKgLimit)
		return rate.FlatRateRM + surchargeKg*rate.PerKgRM

	case domain.RateTypeByDimension:
		// Volumetric weight = L*W*H / divisor; charge the greater of actual vs volumetric
		divisor := rate.VolumetricDivisor
		if divisor == 0 {
			divisor = 5000
		}
		volumetricKg := (req.LengthCM * req.WidthCM * req.HeightCM) / divisor
		chargeableKg := math.Max(req.WeightKg, volumetricKg)
		// Use weight tiers if configured, else flat_plus_kg model
		var tiers []domain.WeightTier
		json.Unmarshal([]byte(rate.WeightTiers), &tiers) //nolint:errcheck
		for _, t := range tiers {
			if chargeableKg >= t.MinKg && (t.MaxKg == 0 || chargeableKg <= t.MaxKg) {
				return t.RateRM
			}
		}
		return rate.FlatRateRM + math.Max(0, chargeableKg-rate.FreeKgLimit)*rate.PerKgRM

	case domain.RateTypeByPrice:
		// Tiered by order subtotal
		var tiers []domain.PriceTier
		json.Unmarshal([]byte(rate.PriceTiers), &tiers) //nolint:errcheck
		for _, t := range tiers {
			if req.OrderRM >= t.MinRM && (t.MaxRM == 0 || req.OrderRM <= t.MaxRM) {
				return t.RateRM
			}
		}
		return rate.FlatRateRM // fallback

	case domain.RateTypeByItem:
		// Per-item fee with optional cap
		fee := float64(req.ItemCount) * rate.PerItemRM
		if rate.MaxItemFee > 0 && fee > rate.MaxItemFee {
			fee = rate.MaxItemFee
		}
		return fee

	case domain.RateTypeCOD:
		// Base flat fee + percentage of order, with optional cap
		fee := rate.CODFeeRM + req.OrderRM*rate.CODFeePercent/100
		if rate.CODMaxFeeRM > 0 && fee > rate.CODMaxFeeRM {
			fee = rate.CODMaxFeeRM
		}
		return fee

	default:
		return rate.FlatRateRM
	}
}
