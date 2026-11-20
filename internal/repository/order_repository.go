package repository

import (
	"fmt"
	"time"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type orderRepository struct{ db *gorm.DB }

func NewOrderRepository(db *gorm.DB) domain.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(o *domain.Order) error {
	return r.db.Create(o).Error
}

func (r *orderRepository) FindByID(id uuid.UUID) (*domain.Order, error) {
	var o domain.Order
	err := r.db.Preload("Items").Preload("User").First(&o, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepository) FindByOrderNo(no string) (*domain.Order, error) {
	var o domain.Order
	err := r.db.Preload("Items").Preload("User").First(&o, "order_no = ?", no).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepository) FindByPaymentRef(ref string) (*domain.Order, error) {
	var o domain.Order
	err := r.db.Preload("Items").Preload("User").First(&o, "payment_ref = ?", ref).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepository) FindByUserID(userID uuid.UUID, page, pageSize int) ([]domain.Order, int64, error) {
	q := r.db.Model(&domain.Order{}).Preload("Items").Where("user_id = ?", userID)
	var total int64
	q.Count(&total)
	var orders []domain.Order
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

func (r *orderRepository) Update(o *domain.Order) error {
	return r.db.Save(o).Error
}

func (r *orderRepository) List(status domain.OrderStatus, search string, page, pageSize int) ([]domain.Order, int64, error) {
	q := r.db.Model(&domain.Order{}).Preload("Items").Preload("User")
	if status != "" {
		q = q.Where("orders.status = ?", status)
	}
	if search != "" {
		like := fmt.Sprintf("%%%s%%", search)
		q = q.Joins("JOIN users ON users.id = orders.user_id").
			Where("orders.order_no ILIKE ? OR users.email ILIKE ? OR users.first_name ILIKE ? OR users.phone ILIKE ?",
				like, like, like, like)
	}
	var total int64
	q.Count(&total)
	var orders []domain.Order
	offset := (page - 1) * pageSize
	err := q.Order("orders.created_at DESC").Offset(offset).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

func (r *orderRepository) CountByStatus(status domain.OrderStatus) (int64, error) {
	var c int64
	err := r.db.Model(&domain.Order{}).Where("status = ?", status).Count(&c).Error
	return c, err
}

func (r *orderRepository) TotalRevenue() (float64, error) {
	var total float64
	err := r.db.Model(&domain.Order{}).
		Where("status IN ?", []domain.OrderStatus{domain.OrderStatusPaid, domain.OrderStatusProcessing, domain.OrderStatusShipped, domain.OrderStatusDelivered}).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&total).Error
	return total, err
}

func (r *orderRepository) RevenueByMonth(months int) ([]domain.MonthlyRevenue, error) {
	type row struct {
		Month   string
		Revenue float64
		Orders  int64
	}
	var rows []row
	err := r.db.Raw(`
		SELECT TO_CHAR(created_at, 'YYYY-MM') AS month,
		       COALESCE(SUM(total_amount), 0) AS revenue,
		       COUNT(*) AS orders
		FROM orders
		WHERE status IN ('paid','processing','shipped','delivered')
		  AND created_at >= NOW() - INTERVAL '? months'
		GROUP BY month
		ORDER BY month ASC
	`, months).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]domain.MonthlyRevenue, len(rows))
	for i, rw := range rows {
		result[i] = domain.MonthlyRevenue{Month: rw.Month, Revenue: rw.Revenue, Orders: rw.Orders}
	}
	return result, nil
}

// ── Cart ──────────────────────────────────────────────────────────────────────

func (r *orderRepository) UpsertCartItem(item *domain.CartItem) error {
	existing, err := r.FindCartItem(item.UserID, item.VariantID)
	if err != nil {
		// Not found — create
		return r.db.Create(item).Error
	}
	existing.Quantity = item.Quantity
	existing.UpdatedAt = time.Now()
	return r.db.Save(existing).Error
}

func (r *orderRepository) GetCartItems(userID uuid.UUID) ([]domain.CartItem, error) {
	var items []domain.CartItem
	err := r.db.Preload("Variant").Preload("Variant.Product").
		Where("user_id = ?", userID).
		Find(&items).Error
	return items, err
}

func (r *orderRepository) DeleteCartItem(id uuid.UUID, userID uuid.UUID) error {
	return r.db.Delete(&domain.CartItem{}, "id = ? AND user_id = ?", id, userID).Error
}

func (r *orderRepository) ClearCart(userID uuid.UUID) error {
	return r.db.Delete(&domain.CartItem{}, "user_id = ?", userID).Error
}

func (r *orderRepository) FindCartItem(userID, variantID uuid.UUID) (*domain.CartItem, error) {
	var item domain.CartItem
	err := r.db.First(&item, "user_id = ? AND variant_id = ?", userID, variantID).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
