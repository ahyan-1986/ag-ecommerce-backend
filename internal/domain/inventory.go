package domain

import (
	"time"

	"github.com/google/uuid"
)

// InventoryMovement records every stock change (receive, sale, adjustment, return).
type MovementType string

const (
	MovementReceive  MovementType = "receive"
	MovementSale     MovementType = "sale"
	MovementAdjust   MovementType = "adjust"
	MovementReturn   MovementType = "return"
	MovementWriteOff MovementType = "write_off"
)

type InventoryMovement struct {
	ID            uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	VariantID     uuid.UUID    `gorm:"type:uuid;not null;index" json:"variantId"`
	Type          MovementType `gorm:"type:varchar(20);not null" json:"type"`
	Quantity      int          `gorm:"not null" json:"quantity"` // positive = in, negative = out
	BalanceBefore int          `gorm:"not null" json:"balanceBefore"`
	BalanceAfter  int          `gorm:"not null" json:"balanceAfter"`
	Reference     string       `gorm:"type:varchar(100);default:''" json:"reference"` // order no, PO no
	Note          string       `gorm:"type:text;default:''" json:"note"`
	CreatedBy     uuid.UUID    `gorm:"type:uuid;not null" json:"createdBy"`
	CreatedAt     time.Time    `gorm:"autoCreateTime" json:"createdAt"`

	Variant *ProductVariant `gorm:"foreignKey:VariantID" json:"variant,omitempty"`
}

func (InventoryMovement) TableName() string { return "inventory_movements" }

// ── Request / Response DTOs ───────────────────────────────────────────────────

type StockAdjustRequest struct {
	VariantID uuid.UUID    `json:"variantId" binding:"required"`
	Type      MovementType `json:"type" binding:"required"`
	Quantity  int          `json:"quantity" binding:"required"`
	Reference string       `json:"reference"`
	Note      string       `json:"note"`
}

type LowStockItem struct {
	VariantID     uuid.UUID `json:"variantId"`
	ProductName   string    `json:"productName"`
	VariantName   string    `json:"variantName"`
	SKU           string    `json:"sku"`
	StockOnHand   int       `json:"stockOnHand"`
	LowStockAlert int       `json:"lowStockAlert"`
}

// BulkReceiveItem is one line in a purchase-order receipt.
type BulkReceiveItem struct {
	VariantID uuid.UUID `json:"variantId" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,min=1"`
	Note      string    `json:"note"`
}

type BulkReceiveRequest struct {
	Reference string            `json:"reference" binding:"required"`
	Note      string            `json:"note"`
	Items     []BulkReceiveItem `json:"items" binding:"required,min=1"`
}

type InventorySummary struct {
	TotalVariants    int64   `json:"totalVariants"`
	TotalStockValue  float64 `json:"totalStockValue"` // SUM(stockOnHand * retailPrice)
	LowStockCount    int64   `json:"lowStockCount"`
	OutOfStockCount  int64   `json:"outOfStockCount"`
	TotalMovements   int64   `json:"totalMovements"`
}

// ── Interfaces ────────────────────────────────────────────────────────────────

type InventoryRepository interface {
	CreateMovement(m *InventoryMovement) error
	ListMovements(variantID *uuid.UUID, movType MovementType, page, pageSize int) ([]InventoryMovement, int64, error)
	GetLowStockItems() ([]LowStockItem, error)
	CountLowStock() (int64, error)
	GetSummary() (*InventorySummary, error)
}

type InventoryService interface {
	AdjustStock(adminID uuid.UUID, req StockAdjustRequest) (*InventoryMovement, error)
	BulkReceive(adminID uuid.UUID, req BulkReceiveRequest) ([]InventoryMovement, error)
	ListMovements(variantID *uuid.UUID, movType MovementType, page, pageSize int) ([]InventoryMovement, int64, error)
	GetLowStockItems() ([]LowStockItem, error)
	GetSummary() (*InventorySummary, error)
}
