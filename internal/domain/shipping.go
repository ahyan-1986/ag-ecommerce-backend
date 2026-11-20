package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ── Shipping Zone ─────────────────────────────────────────────────────────────
// A zone groups countries/states/postcodes that share the same rates.

type ShippingZone struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`  // e.g. "West Malaysia"
	Description string         `gorm:"type:text;default:''" json:"description"` // optional notes
	Countries   string         `gorm:"type:text;default:'[]'" json:"countries"` // JSON []string  e.g. ["MY"]
	States      string         `gorm:"type:text;default:'[]'" json:"states"`    // JSON []string  e.g. ["Selangor","KL"]
	IsActive    bool           `gorm:"not null;default:true" json:"isActive"`
	SortOrder   int            `gorm:"default:0" json:"sortOrder"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Rates []ShippingRate `gorm:"foreignKey:ZoneID" json:"rates,omitempty"`
}

func (ShippingZone) TableName() string { return "shipping_zones" }

// ── Shipping Rate ────────────────────────────────────────────────────────────
// A rate within a zone. Supports multiple pricing models used in Asian e-commerce.

type ShippingRateType string

const (
	// ── Basic ────────────────────────────────────────────────────────────
	RateTypeFlat     ShippingRateType = "flat"       // fixed fee regardless of weight/quantity
	RateTypeFree     ShippingRateType = "free"        // free; optionally conditioned on minOrderRM
	// ── Weight-based ────────────────────────────────────────────────────
	RateTypeByWeight    ShippingRateType = "by_weight"     // tiered by total order weight (kg)
	RateTypeFlatPlusKg  ShippingRateType = "flat_plus_kg"  // flat base fee + per-kg surcharge above threshold
	RateTypeByDimension ShippingRateType = "by_dimension"  // volumetric: (L×W×H)/divisor = chargeable weight
	// ── Value-based ─────────────────────────────────────────────────────
	RateTypeByPrice ShippingRateType = "by_price" // tiered by order subtotal (RM)
	// ── Quantity-based ──────────────────────────────────────────────────
	RateTypeByItem ShippingRateType = "by_item" // fixed fee per item/unit in order
	// ── Special ─────────────────────────────────────────────────────────
	RateTypeCOD ShippingRateType = "cod" // cash on delivery; charges COD handling fee
)

type ShippingRate struct {
	ID          uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ZoneID      uuid.UUID        `gorm:"type:uuid;not null;index" json:"zoneId"`
	Name        string           `gorm:"type:varchar(100);not null" json:"name"`
	Carrier     string           `gorm:"type:varchar(100);default:''" json:"carrier"`
	RateType    ShippingRateType `gorm:"type:varchar(20);not null" json:"rateType"`

	// ── Flat / base ──────────────────────────────────────────────────────
	FlatRateRM float64 `gorm:"type:numeric(12,2);default:0" json:"flatRateRM"` // flat or base fee

	// ── Weight-based ─────────────────────────────────────────────────────
	WeightTiers string  `gorm:"type:text;default:'[]'" json:"weightTiers"` // JSON []WeightTier (by_weight)
	PerKgRM     float64 `gorm:"type:numeric(12,2);default:0" json:"perKgRM"` // per-kg surcharge (flat_plus_kg)
	FreeKgLimit float64 `gorm:"type:numeric(12,2);default:0" json:"freeKgLimit"` // first N kg free in flat_plus_kg

	// ── Volumetric ────────────────────────────────────────────────────────
	VolumetricDivisor float64 `gorm:"type:numeric(12,4);default:5000" json:"volumetricDivisor"` // cm³/divisor = kg (default 5000)

	// ── Price-tiered ──────────────────────────────────────────────────────
	PriceTiers string `gorm:"type:text;default:'[]'" json:"priceTiers"` // JSON []PriceTier (by_price)

	// ── Per-item ─────────────────────────────────────────────────────────
	PerItemRM  float64 `gorm:"type:numeric(12,2);default:0" json:"perItemRM"`  // fee per item
	MaxItemFee float64 `gorm:"type:numeric(12,2);default:0" json:"maxItemFee"` // cap on item fee (0=no cap)

	// ── COD ───────────────────────────────────────────────────────────────
	CODFeeRM      float64 `gorm:"type:numeric(12,2);default:0" json:"codFeeRM"`      // COD handling flat fee
	CODFeePercent float64 `gorm:"type:numeric(6,4);default:0" json:"codFeePercent"`  // COD % of order value
	CODMaxFeeRM   float64 `gorm:"type:numeric(12,2);default:0" json:"codMaxFeeRM"`   // COD fee cap (0=no cap)

	// ── Conditions ────────────────────────────────────────────────────────
	MinOrderRM float64 `gorm:"type:numeric(12,2);default:0" json:"minOrderRM"` // minimum order to use this rate
	MaxOrderRM float64 `gorm:"type:numeric(12,2);default:0" json:"maxOrderRM"` // maximum order (0=no max)

	// ── Delivery estimate ─────────────────────────────────────────────────
	EstDaysMin int    `gorm:"default:0" json:"estDaysMin"` // estimated min delivery days
	EstDaysMax int    `gorm:"default:0" json:"estDaysMax"` // estimated max delivery days
	CutoffTime string `gorm:"type:varchar(10);default:''" json:"cutoffTime"` // e.g. "14:00" same-day cutoff

	IsActive  bool           `gorm:"not null;default:true" json:"isActive"`
	SortOrder int            `gorm:"default:0" json:"sortOrder"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Zone *ShippingZone `gorm:"foreignKey:ZoneID" json:"zone,omitempty"`
}

func (ShippingRate) TableName() string { return "shipping_rates" }

// WeightTier is embedded as JSON inside ShippingRate.WeightTiers (by_weight)
type WeightTier struct {
	MinKg  float64 `json:"minKg"`
	MaxKg  float64 `json:"maxKg"` // 0 = unlimited
	RateRM float64 `json:"rateRM"`
}

// PriceTier is embedded as JSON inside ShippingRate.PriceTiers (by_price)
type PriceTier struct {
	MinRM  float64 `json:"minRM"`
	MaxRM  float64 `json:"maxRM"` // 0 = unlimited
	RateRM float64 `json:"rateRM"`
}

// ── DTOs ─────────────────────────────────────────────────────────────────────

type CreateShippingZoneRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Countries   []string `json:"countries"`
	States      []string `json:"states"`
	IsActive    bool     `json:"isActive"`
	SortOrder   int      `json:"sortOrder"`
}

type UpdateShippingZoneRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Countries   []string `json:"countries"`
	States      []string `json:"states"`
	IsActive    *bool    `json:"isActive"`
	SortOrder   *int     `json:"sortOrder"`
}

type CreateShippingRateRequest struct {
	ZoneID    uuid.UUID        `json:"zoneId" binding:"required"`
	Name      string           `json:"name" binding:"required"`
	Carrier   string           `json:"carrier"`
	RateType  ShippingRateType `json:"rateType" binding:"required"`
	IsActive  bool             `json:"isActive"`
	SortOrder int              `json:"sortOrder"`

	// Flat / base
	FlatRateRM float64 `json:"flatRateRM"`

	// Weight-based
	WeightTiers []WeightTier `json:"weightTiers"`
	PerKgRM     float64      `json:"perKgRM"`
	FreeKgLimit float64      `json:"freeKgLimit"`

	// Volumetric
	VolumetricDivisor float64 `json:"volumetricDivisor"`

	// Price-tiered
	PriceTiers []PriceTier `json:"priceTiers"`

	// Per-item
	PerItemRM  float64 `json:"perItemRM"`
	MaxItemFee float64 `json:"maxItemFee"`

	// COD
	CODFeeRM      float64 `json:"codFeeRM"`
	CODFeePercent float64 `json:"codFeePercent"`
	CODMaxFeeRM   float64 `json:"codMaxFeeRM"`

	// Conditions
	MinOrderRM float64 `json:"minOrderRM"`
	MaxOrderRM float64 `json:"maxOrderRM"`

	// Delivery estimate
	EstDaysMin int    `json:"estDaysMin"`
	EstDaysMax int    `json:"estDaysMax"`
	CutoffTime string `json:"cutoffTime"`
}

type UpdateShippingRateRequest struct {
	Name     *string           `json:"name"`
	Carrier  *string           `json:"carrier"`
	RateType *ShippingRateType `json:"rateType"`
	IsActive *bool             `json:"isActive"`
	SortOrder *int             `json:"sortOrder"`

	// Flat / base
	FlatRateRM *float64 `json:"flatRateRM"`

	// Weight-based
	WeightTiers []WeightTier `json:"weightTiers"`
	PerKgRM     *float64     `json:"perKgRM"`
	FreeKgLimit *float64     `json:"freeKgLimit"`

	// Volumetric
	VolumetricDivisor *float64 `json:"volumetricDivisor"`

	// Price-tiered
	PriceTiers []PriceTier `json:"priceTiers"`

	// Per-item
	PerItemRM  *float64 `json:"perItemRM"`
	MaxItemFee *float64 `json:"maxItemFee"`

	// COD
	CODFeeRM      *float64 `json:"codFeeRM"`
	CODFeePercent *float64 `json:"codFeePercent"`
	CODMaxFeeRM   *float64 `json:"codMaxFeeRM"`

	// Conditions
	MinOrderRM *float64 `json:"minOrderRM"`
	MaxOrderRM *float64 `json:"maxOrderRM"`

	// Delivery estimate
	EstDaysMin *int    `json:"estDaysMin"`
	EstDaysMax *int    `json:"estDaysMax"`
	CutoffTime *string `json:"cutoffTime"`
}

// CalculateShippingRequest is used to get applicable rates at checkout.
type CalculateShippingRequest struct {
	Country    string  `json:"country" binding:"required"` // e.g. "MY"
	State      string  `json:"state"`                      // e.g. "Selangor"
	WeightKg   float64 `json:"weightKg"`                   // total order weight
	OrderRM    float64 `json:"orderRM"`                    // order subtotal
	ItemCount  int     `json:"itemCount"`                  // number of items (for by_item)
	// Volumetric inputs (cm)
	LengthCM float64 `json:"lengthCm"`
	WidthCM  float64 `json:"widthCm"`
	HeightCM float64 `json:"heightCm"`
}

// CalculatedRate is returned to the customer showing available shipping options.
type CalculatedRate struct {
	RateID     uuid.UUID `json:"rateId"`
	ZoneName   string    `json:"zoneName"`
	Name       string    `json:"name"`
	Carrier    string    `json:"carrier"`
	RateType   string    `json:"rateType"`
	CostRM     float64   `json:"costRM"`
	EstDaysMin int       `json:"estDaysMin"`
	EstDaysMax int       `json:"estDaysMax"`
	CutoffTime string    `json:"cutoffTime"`
}

// ── Repository & Service Interfaces ─────────────────────────────────────────

type ShippingRepository interface {
	ListZones() ([]ShippingZone, error)
	GetZone(id uuid.UUID) (*ShippingZone, error)
	CreateZone(z *ShippingZone) error
	UpdateZone(z *ShippingZone) error
	DeleteZone(id uuid.UUID) error

	ListRates(zoneID *uuid.UUID) ([]ShippingRate, error)
	GetRate(id uuid.UUID) (*ShippingRate, error)
	CreateRate(r *ShippingRate) error
	UpdateRate(r *ShippingRate) error
	DeleteRate(id uuid.UUID) error
}

type ShippingService interface {
	ListZones() ([]ShippingZone, error)
	CreateZone(req CreateShippingZoneRequest) (*ShippingZone, error)
	UpdateZone(id uuid.UUID, req UpdateShippingZoneRequest) (*ShippingZone, error)
	DeleteZone(id uuid.UUID) error

	ListRates(zoneID *uuid.UUID) ([]ShippingRate, error)
	GetRate(id uuid.UUID) (*ShippingRate, error)
	CreateRate(req CreateShippingRateRequest) (*ShippingRate, error)
	UpdateRate(id uuid.UUID, req UpdateShippingRateRequest) (*ShippingRate, error)
	DeleteRate(id uuid.UUID) error

	CalculateRates(req CalculateShippingRequest) ([]CalculatedRate, error)
}
