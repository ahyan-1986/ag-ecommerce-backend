package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductStatus controls product visibility in the shop.
type ProductStatus string

const (
	ProductStatusDraft     ProductStatus = "draft"
	ProductStatusPublished ProductStatus = "published"
	ProductStatusArchived  ProductStatus = "archived"
)

// Product represents a cat litter product sold on AsianMum.
type Product struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SKU         string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"sku"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	NameZH      string         `gorm:"type:varchar(255);default:''" json:"nameZh"` // Chinese name
	Description string         `gorm:"type:text;default:''" json:"description"`
	Category    string         `gorm:"type:varchar(100);default:''" json:"category"`
	Brand       string         `gorm:"type:varchar(100);default:'AsianMum'" json:"brand"`
	ImageURL    string         `gorm:"type:varchar(500);default:''" json:"imageUrl"`
	ImageURLs   string         `gorm:"type:text;default:'[]'" json:"imageUrls"` // JSON array of additional images
	Tags        string         `gorm:"type:text;default:'[]'" json:"tags"`      // JSON array
	Status      ProductStatus  `gorm:"type:varchar(20);not null;default:'draft'" json:"status"`
	IsFeatured  bool           `gorm:"not null;default:false" json:"isFeatured"`
	SortOrder   int            `gorm:"default:0" json:"sortOrder"`
	CreatedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"createdBy"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Variants []ProductVariant `gorm:"foreignKey:ProductID" json:"variants,omitempty"`
}

func (Product) TableName() string { return "products" }

// ProductVariant is a specific SKU variant (e.g., 5kg, 10kg, scented, unscented).
type ProductVariant struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProductID  uuid.UUID `gorm:"type:uuid;not null;index" json:"productId"`
	Name       string    `gorm:"type:varchar(150);not null" json:"name"` // e.g. "5kg - Lavender"
	SKU        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"sku"`
	Attributes string    `gorm:"type:text;default:'{}'" json:"attributes"` // JSON: weight, scent, type
	ImageURL   string    `gorm:"type:varchar(500);default:''" json:"imageUrl"`
	// Retail price (customer)
	RetailPrice float64 `gorm:"type:numeric(12,2);not null;default:0" json:"retailPrice"`
	// Dealer price (lower)
	DealerPrice float64 `gorm:"type:numeric(12,2);not null;default:0" json:"dealerPrice"`
	BV          float64 `gorm:"type:numeric(12,2);not null;default:0" json:"bv"`
	PV          float64 `gorm:"type:numeric(12,2);not null;default:0" json:"pv"`
	// Stock
	StockOnHand   int            `gorm:"not null;default:0" json:"stockOnHand"`
	StockReserved int            `gorm:"not null;default:0" json:"stockReserved"`
	LowStockAlert int            `gorm:"not null;default:10" json:"lowStockAlert"`
	Weight        float64        `gorm:"type:numeric(8,3);default:0" json:"weight"` // kg, for shipping
	IsActive      bool           `gorm:"not null;default:true" json:"isActive"`
	SortOrder     int            `gorm:"default:0" json:"sortOrder"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (ProductVariant) TableName() string { return "product_variants" }

func (v *ProductVariant) AvailableStock() int {
	avail := v.StockOnHand - v.StockReserved
	if avail < 0 {
		return 0
	}
	return avail
}

// ProductReview is a customer review for a product.
type ProductReview struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;index" json:"productId"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	OrderID   uuid.UUID `gorm:"type:uuid;not null" json:"orderId"`
	Rating    int       `gorm:"not null" json:"rating"` // 1-5
	Title     string    `gorm:"type:varchar(255);default:''" json:"title"`
	Body      string    `gorm:"type:text;default:''" json:"body"`
	IsVisible bool      `gorm:"not null;default:true" json:"isVisible"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (ProductReview) TableName() string { return "product_reviews" }

// ── Request / Response DTOs ───────────────────────────────────────────────────

type CreateProductRequest struct {
	Name        string `json:"name" binding:"required"`
	NameZH      string `json:"nameZh"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Brand       string `json:"brand"`
	Tags        string `json:"tags"`
	Status      string `json:"status"`
	IsFeatured  bool   `json:"isFeatured"`
	SortOrder   int    `json:"sortOrder"`
}

type UpdateProductRequest struct {
	Name        *string `json:"name"`
	NameZH      *string `json:"nameZh"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	Brand       *string `json:"brand"`
	Tags        *string `json:"tags"`
	Status      *string `json:"status"`
	IsFeatured  *bool   `json:"isFeatured"`
	SortOrder   *int    `json:"sortOrder"`
}

type CreateVariantRequest struct {
	Name          string  `json:"name" binding:"required"`
	SKU           string  `json:"sku" binding:"required"`
	Attributes    string  `json:"attributes"`
	RetailPrice   float64 `json:"retailPrice"`
	DealerPrice   float64 `json:"dealerPrice"`
	BV            float64 `json:"bv"`
	PV            float64 `json:"pv"`
	StockOnHand   int     `json:"stockOnHand"`
	LowStockAlert int     `json:"lowStockAlert"`
	Weight        float64 `json:"weight"`
	IsActive      bool    `json:"isActive"`
	SortOrder     int     `json:"sortOrder"`
}

type UpdateVariantRequest struct {
	Name          *string  `json:"name"`
	Attributes    *string  `json:"attributes"`
	RetailPrice   *float64 `json:"retailPrice"`
	DealerPrice   *float64 `json:"dealerPrice"`
	BV            *float64 `json:"bv"`
	PV            *float64 `json:"pv"`
	LowStockAlert *int     `json:"lowStockAlert"`
	Weight        *float64 `json:"weight"`
	IsActive      *bool    `json:"isActive"`
	SortOrder     *int     `json:"sortOrder"`
}

type CreateReviewRequest struct {
	ProductID uuid.UUID `json:"productId" binding:"required"`
	OrderID   uuid.UUID `json:"orderId" binding:"required"`
	Rating    int       `json:"rating" binding:"required,min=1,max=5"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
}

// ProductListResponse wraps product list with pagination.
type ProductListResponse struct {
	Items    []Product `json:"items"`
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"pageSize"`
}

// ── Product Repository & Service interfaces ───────────────────────────────────

type ProductRepository interface {
	Create(p *Product) error
	FindByID(id uuid.UUID) (*Product, error)
	FindBySKU(sku string) (*Product, error)
	Update(p *Product) error
	Delete(id uuid.UUID) error
	List(status ProductStatus, category, search string, page, pageSize int) ([]Product, int64, error)
	ListPublished(category, search string, page, pageSize int) ([]Product, int64, error)
	ListFeatured() ([]Product, error)

	// Variants
	CreateVariant(v *ProductVariant) error
	FindVariantByID(id uuid.UUID) (*ProductVariant, error)
	UpdateVariant(v *ProductVariant) error
	DeleteVariant(id uuid.UUID) error
	ListVariantsByProduct(productID uuid.UUID) ([]ProductVariant, error)

	// Reviews
	CreateReview(r *ProductReview) error
	ListReviews(productID uuid.UUID, page, pageSize int) ([]ProductReview, int64, error)
	HasReviewed(userID, orderID, productID uuid.UUID) (bool, error)
}

type ProductService interface {
	CreateProduct(adminID uuid.UUID, req CreateProductRequest) (*Product, error)
	GetProduct(id uuid.UUID) (*Product, error)
	UpdateProduct(id uuid.UUID, req UpdateProductRequest) (*Product, error)
	DeleteProduct(id uuid.UUID) error
	ListProducts(status ProductStatus, category, search string, page, pageSize int) ([]Product, int64, error)
	ListPublishedProducts(category, search string, page, pageSize int) ([]Product, int64, error)
	ListFeaturedProducts() ([]Product, error)
	UploadProductImage(productID uuid.UUID, variantID *uuid.UUID, data []byte, contentType string) (string, error)

	CreateVariant(productID uuid.UUID, req CreateVariantRequest) (*ProductVariant, error)
	UpdateVariant(variantID uuid.UUID, req UpdateVariantRequest) (*ProductVariant, error)
	DeleteVariant(variantID uuid.UUID) error

	CreateReview(userID uuid.UUID, req CreateReviewRequest) (*ProductReview, error)
	ListReviews(productID uuid.UUID, page, pageSize int) ([]ProductReview, int64, error)
}
