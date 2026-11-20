package repository

import (
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type inventoryRepository struct{ db *gorm.DB }

func NewInventoryRepository(db *gorm.DB) domain.InventoryRepository {
	return &inventoryRepository{db: db}
}

func (r *inventoryRepository) CreateMovement(m *domain.InventoryMovement) error {
	return r.db.Create(m).Error
}

func (r *inventoryRepository) ListMovements(variantID *uuid.UUID, movType domain.MovementType, page, pageSize int) ([]domain.InventoryMovement, int64, error) {
	q := r.db.Model(&domain.InventoryMovement{}).
		Preload("Variant").
		Preload("Variant.Product")
	if variantID != nil {
		q = q.Where("variant_id = ?", *variantID)
	}
	if movType != "" {
		q = q.Where("type = ?", movType)
	}
	var total int64
	q.Count(&total)
	var items []domain.InventoryMovement
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *inventoryRepository) GetLowStockItems() ([]domain.LowStockItem, error) {
	var items []domain.LowStockItem
	err := r.db.Raw(`
		SELECT
			pv.id AS variant_id,
			p.name AS product_name,
			pv.name AS variant_name,
			pv.sku,
			pv.stock_on_hand,
			pv.low_stock_alert
		FROM product_variants pv
		JOIN products p ON p.id = pv.product_id AND p.deleted_at IS NULL
		WHERE pv.deleted_at IS NULL
		  AND pv.is_active = true
		  AND pv.stock_on_hand <= pv.low_stock_alert
		ORDER BY pv.stock_on_hand ASC
	`).Scan(&items).Error
	return items, err
}

func (r *inventoryRepository) CountLowStock() (int64, error) {
	var c int64
	err := r.db.Raw(`
		SELECT COUNT(*) FROM product_variants pv
		JOIN products p ON p.id = pv.product_id AND p.deleted_at IS NULL
		WHERE pv.deleted_at IS NULL AND pv.is_active = true
		  AND pv.stock_on_hand <= pv.low_stock_alert
	`).Scan(&c).Error
	return c, err
}

func (r *inventoryRepository) GetSummary() (*domain.InventorySummary, error) {
	var s domain.InventorySummary
	err := r.db.Raw(`
		SELECT
			COUNT(pv.id)                                            AS total_variants,
			COALESCE(SUM(pv.stock_on_hand * pv.retail_price), 0)  AS total_stock_value,
			COUNT(*) FILTER (WHERE pv.stock_on_hand <= pv.low_stock_alert AND pv.stock_on_hand > 0) AS low_stock_count,
			COUNT(*) FILTER (WHERE pv.stock_on_hand = 0)           AS out_of_stock_count
		FROM product_variants pv
		JOIN products p ON p.id = pv.product_id AND p.deleted_at IS NULL
		WHERE pv.deleted_at IS NULL AND pv.is_active = true
	`).Scan(&s).Error
	if err != nil {
		return nil, err
	}
	r.db.Model(&domain.InventoryMovement{}).Count(&s.TotalMovements)
	return &s, nil
}
