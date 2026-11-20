package repository

import (
	"fmt"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type productRepository struct{ db *gorm.DB }

func NewProductRepository(db *gorm.DB) domain.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(p *domain.Product) error {
	return r.db.Create(p).Error
}

func (r *productRepository) FindByID(id uuid.UUID) (*domain.Product, error) {
	var p domain.Product
	err := r.db.Preload("Variants").First(&p, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *productRepository) FindBySKU(sku string) (*domain.Product, error) {
	var p domain.Product
	err := r.db.Preload("Variants").First(&p, "sku = ?", sku).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *productRepository) Update(p *domain.Product) error {
	return r.db.Save(p).Error
}

func (r *productRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Product{}, "id = ?", id).Error
}

func (r *productRepository) List(status domain.ProductStatus, category, search string, page, pageSize int) ([]domain.Product, int64, error) {
	q := r.db.Model(&domain.Product{}).Preload("Variants")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if search != "" {
		like := fmt.Sprintf("%%%s%%", search)
		q = q.Where("name ILIKE ? OR name_z_h ILIKE ? OR sku ILIKE ?", like, like, like)
	}
	var total int64
	q.Count(&total)
	var products []domain.Product
	offset := (page - 1) * pageSize
	err := q.Order("sort_order ASC, created_at DESC").Offset(offset).Limit(pageSize).Find(&products).Error
	return products, total, err
}

func (r *productRepository) ListPublished(category, search string, page, pageSize int) ([]domain.Product, int64, error) {
	return r.List(domain.ProductStatusPublished, category, search, page, pageSize)
}

func (r *productRepository) ListFeatured() ([]domain.Product, error) {
	var products []domain.Product
	err := r.db.Preload("Variants").
		Where("status = ? AND is_featured = true", domain.ProductStatusPublished).
		Order("sort_order ASC").
		Find(&products).Error
	return products, err
}

func (r *productRepository) CreateVariant(v *domain.ProductVariant) error {
	return r.db.Create(v).Error
}

func (r *productRepository) FindVariantByID(id uuid.UUID) (*domain.ProductVariant, error) {
	var v domain.ProductVariant
	err := r.db.First(&v, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *productRepository) UpdateVariant(v *domain.ProductVariant) error {
	return r.db.Save(v).Error
}

func (r *productRepository) DeleteVariant(id uuid.UUID) error {
	return r.db.Delete(&domain.ProductVariant{}, "id = ?", id).Error
}

func (r *productRepository) ListVariantsByProduct(productID uuid.UUID) ([]domain.ProductVariant, error) {
	var variants []domain.ProductVariant
	err := r.db.Where("product_id = ?", productID).Order("sort_order ASC").Find(&variants).Error
	return variants, err
}

func (r *productRepository) CreateReview(rv *domain.ProductReview) error {
	return r.db.Create(rv).Error
}

func (r *productRepository) ListReviews(productID uuid.UUID, page, pageSize int) ([]domain.ProductReview, int64, error) {
	q := r.db.Model(&domain.ProductReview{}).
		Preload("User").
		Where("product_id = ? AND is_visible = true", productID)
	var total int64
	q.Count(&total)
	var reviews []domain.ProductReview
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&reviews).Error
	return reviews, total, err
}

func (r *productRepository) HasReviewed(userID, orderID, productID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&domain.ProductReview{}).
		Where("user_id = ? AND order_id = ? AND product_id = ?", userID, orderID, productID).
		Count(&count).Error
	return count > 0, err
}
