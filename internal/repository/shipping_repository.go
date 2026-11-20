package repository

import (
	"encoding/json"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type shippingRepository struct{ db *gorm.DB }

func NewShippingRepository(db *gorm.DB) domain.ShippingRepository {
	return &shippingRepository{db: db}
}

// ── Zones ─────────────────────────────────────────────────────────────────────

func (r *shippingRepository) ListZones() ([]domain.ShippingZone, error) {
	var zones []domain.ShippingZone
	err := r.db.Preload("Rates", "deleted_at IS NULL AND is_active = true").
		Where("deleted_at IS NULL").
		Order("sort_order ASC, created_at ASC").
		Find(&zones).Error
	return zones, err
}

func (r *shippingRepository) GetZone(id uuid.UUID) (*domain.ShippingZone, error) {
	var z domain.ShippingZone
	err := r.db.Preload("Rates", "deleted_at IS NULL").Where("id = ?", id).First(&z).Error
	return &z, err
}

func (r *shippingRepository) CreateZone(z *domain.ShippingZone) error {
	return r.db.Create(z).Error
}

func (r *shippingRepository) UpdateZone(z *domain.ShippingZone) error {
	return r.db.Save(z).Error
}

func (r *shippingRepository) DeleteZone(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&domain.ShippingZone{}).Error
}

// ── Rates ─────────────────────────────────────────────────────────────────────

func (r *shippingRepository) ListRates(zoneID *uuid.UUID) ([]domain.ShippingRate, error) {
	var rates []domain.ShippingRate
	q := r.db.Preload("Zone").Where("deleted_at IS NULL")
	if zoneID != nil {
		q = q.Where("zone_id = ?", *zoneID)
	}
	err := q.Order("sort_order ASC, created_at ASC").Find(&rates).Error
	return rates, err
}

func (r *shippingRepository) GetRate(id uuid.UUID) (*domain.ShippingRate, error) {
	var rate domain.ShippingRate
	err := r.db.Preload("Zone").Where("id = ?", id).First(&rate).Error
	return &rate, err
}

func (r *shippingRepository) CreateRate(rate *domain.ShippingRate) error {
	return r.db.Create(rate).Error
}

func (r *shippingRepository) UpdateRate(rate *domain.ShippingRate) error {
	return r.db.Save(rate).Error
}

func (r *shippingRepository) DeleteRate(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&domain.ShippingRate{}).Error
}

// ── Helper: AutoMigrate ───────────────────────────────────────────────────────

func MigrateShipping(db *gorm.DB) error {
	return db.AutoMigrate(&domain.ShippingZone{}, &domain.ShippingRate{})
}

// jsonEncodeStringSlice marshals a []string to JSON, returning "[]" on nil.
func jsonEncodeStringSlice(ss []string) string {
	if ss == nil {
		return "[]"
	}
	b, _ := json.Marshal(ss)
	return string(b)
}

// jsonDecodeStringSlice parses a JSON string slice from a raw string.
func jsonDecodeStringSlice(raw string) []string {
	if raw == "" || raw == "null" {
		return []string{}
	}
	var ss []string
	json.Unmarshal([]byte(raw), &ss) //nolint:errcheck
	return ss
}
