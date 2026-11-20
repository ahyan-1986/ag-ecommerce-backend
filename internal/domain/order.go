package domain

import (
	"time"

	"github.com/google/uuid"
)

// ── Enums ─────────────────────────────────────────────────────────────────────

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusPaid       OrderStatus = "paid"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

type ShippingStatus string

const (
	ShippingNew       ShippingStatus = "new"
	ShippingPicking   ShippingStatus = "picking"
	ShippingPacked    ShippingStatus = "packed"
	ShippingShipped   ShippingStatus = "shipped"
	ShippingDelivered ShippingStatus = "delivered"
	ShippingCancelled ShippingStatus = "cancelled"
)

// ── Order ─────────────────────────────────────────────────────────────────────

type Order struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderNo     string      `gorm:"type:varchar(30);uniqueIndex;not null" json:"orderNo"` // AG-2026-00001
	UserID      uuid.UUID   `gorm:"type:uuid;not null;index" json:"userId"`
	BuyerType   string      `gorm:"type:varchar(5);not null;default:'M'" json:"buyerType"` // M = Member, G = Guest
	GuestEmail  string      `gorm:"type:varchar(255);default:''" json:"guestEmail"`
	GuestName   string      `gorm:"type:varchar(150);default:''" json:"guestName"`
	GuestPhone  string      `gorm:"type:varchar(30);default:''" json:"guestPhone"`
	SponsorCode string      `gorm:"type:varchar(30);default:''" json:"sponsorCode"`
	Status      OrderStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	// Pricing & Loyalty Points
	SubTotal        float64 `gorm:"type:numeric(12,2);not null;default:0" json:"subTotal"`
	ShippingFee     float64 `gorm:"type:numeric(12,2);not null;default:0" json:"shippingFee"`
	DiscountAmount  float64 `gorm:"type:numeric(12,2);not null;default:0" json:"discountAmount"`
	PointsUsed      float64 `gorm:"type:numeric(12,2);not null;default:0" json:"pointsUsed"`
	PointsDiscount  float64 `gorm:"type:numeric(12,2);not null;default:0" json:"pointsDiscount"`
	TotalAmount     float64 `gorm:"type:numeric(12,2);not null;default:0" json:"totalAmount"`
	TotalBV         float64 `gorm:"type:numeric(12,2);not null;default:0" json:"totalBv"`
	GuestPointsHeld float64 `gorm:"type:numeric(12,2);not null;default:0" json:"guestPointsHeld"`
	GuestPointsPaid bool    `gorm:"not null;default:false" json:"guestPointsPaid"`
	Currency        string  `gorm:"type:varchar(10);not null;default:'MYR'" json:"currency"`
	// Voucher
	VoucherCode string     `gorm:"type:varchar(50);default:''" json:"voucherCode"`
	VoucherID   *uuid.UUID `gorm:"type:uuid" json:"voucherId,omitempty"`
	// Payment
	PaymentMethod string     `gorm:"type:varchar(50);default:''" json:"paymentMethod"` // stripe, cod
	PaymentRef    string     `gorm:"type:varchar(255);default:''" json:"paymentRef"`   // Stripe payment intent ID
	PaidAt        *time.Time `gorm:"type:timestamptz" json:"paidAt,omitempty"`
	// Shipping
	ShippingName     string `gorm:"type:varchar(100);default:''" json:"shippingName"`
	ShippingPhone    string `gorm:"type:varchar(30);default:''" json:"shippingPhone"`
	ShippingLine1    string `gorm:"type:varchar(255);default:''" json:"shippingLine1"`
	ShippingLine2    string `gorm:"type:varchar(255);default:''" json:"shippingLine2"`
	ShippingCity     string `gorm:"type:varchar(100);default:''" json:"shippingCity"`
	ShippingState    string `gorm:"type:varchar(100);default:''" json:"shippingState"`
	ShippingPostcode string `gorm:"type:varchar(20);default:''" json:"shippingPostcode"`
	ShippingCountry  string `gorm:"type:varchar(100);not null;default:'Malaysia'" json:"shippingCountry"`
	// Fulfillment
	ShippingStatus ShippingStatus `gorm:"type:varchar(20);not null;default:'new'" json:"shippingStatus"`
	TrackingNumber string         `gorm:"type:varchar(100);default:''" json:"trackingNumber"`
	Courier        string         `gorm:"type:varchar(100);default:''" json:"courier"`
	Notes          string         `gorm:"type:text;default:''" json:"notes"`
	// Affiliate
	AffiliateEarned bool      `gorm:"not null;default:false" json:"affiliateEarned"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	Items []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	User  *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Order) TableName() string { return "orders" }

type OrderItem struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderID     uuid.UUID `gorm:"type:uuid;not null;index" json:"orderId"`
	ProductID   uuid.UUID `gorm:"type:uuid;not null" json:"productId"`
	VariantID   uuid.UUID `gorm:"type:uuid;not null" json:"variantId"`
	ProductName string    `gorm:"type:varchar(255);not null" json:"productName"`
	VariantName string    `gorm:"type:varchar(150);not null" json:"variantName"`
	SKU         string    `gorm:"type:varchar(100);not null" json:"sku"`
	ImageURL    string    `gorm:"type:varchar(500);default:''" json:"imageUrl"`
	Quantity    int       `gorm:"not null;default:1" json:"quantity"`
	UnitPrice   float64   `gorm:"type:numeric(12,2);not null" json:"unitPrice"`
	TotalPrice  float64   `gorm:"type:numeric(12,2);not null" json:"totalPrice"`
	BV          float64   `gorm:"type:numeric(12,2);not null;default:0" json:"bv"`
	IsFreeGift  bool      `gorm:"not null;default:false" json:"isFreeGift"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (OrderItem) TableName() string { return "order_items" }

// ── Cart (server-side session, optional) ─────────────────────────────────────

type CartItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	VariantID uuid.UUID `gorm:"type:uuid;not null" json:"variantId"`
	Quantity  int       `gorm:"not null;default:1" json:"quantity"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	Variant *ProductVariant `gorm:"foreignKey:VariantID" json:"variant,omitempty"`
}

func (CartItem) TableName() string { return "cart_items" }

// ── Request / Response DTOs ───────────────────────────────────────────────────

type CartItemRequest struct {
	VariantID uuid.UUID `json:"variantId" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,min=1"`
}

type CheckoutRequest struct {
	VoucherCode    string  `json:"voucherCode"`
	PointsToUse    float64 `json:"pointsToUse"`
	SponsorCode    string  `json:"sponsorCode"`
	PaymentMethod  string  `json:"paymentMethod" binding:"required"` // stripe, cod, points
	ShippingRateID string  `json:"shippingRateId"`
	// Shipping
	ShippingName     string `json:"shippingName" binding:"required"`
	ShippingPhone    string `json:"shippingPhone" binding:"required"`
	ShippingLine1    string `json:"shippingLine1" binding:"required"`
	ShippingLine2    string `json:"shippingLine2"`
	ShippingCity     string `json:"shippingCity" binding:"required"`
	ShippingState    string `json:"shippingState" binding:"required"`
	ShippingPostcode string `json:"shippingPostcode" binding:"required"`
	ShippingCountry  string `json:"shippingCountry"`
	Notes            string `json:"notes"`
}

type GuestCheckoutItem struct {
	VariantID uuid.UUID `json:"variantId" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,min=1"`
}

type GuestCheckoutRequest struct {
	GuestEmail       string              `json:"guestEmail" binding:"required,email"`
	GuestName        string              `json:"guestName" binding:"required"`
	GuestPhone       string              `json:"guestPhone" binding:"required"`
	SponsorCode      string              `json:"sponsorCode"`
	Items            []GuestCheckoutItem `json:"items" binding:"required,min=1"`
	PaymentMethod    string              `json:"paymentMethod" binding:"required"`
	ShippingRateID   string              `json:"shippingRateId"`
	ShippingLine1    string              `json:"shippingLine1" binding:"required"`
	ShippingLine2    string              `json:"shippingLine2"`
	ShippingCity     string              `json:"shippingCity" binding:"required"`
	ShippingState    string              `json:"shippingState" binding:"required"`
	ShippingPostcode string              `json:"shippingPostcode" binding:"required"`
	ShippingCountry  string              `json:"shippingCountry"`
	Notes            string              `json:"notes"`
}

type UpdateFulfillmentRequest struct {
	ShippingStatus ShippingStatus `json:"shippingStatus" binding:"required"`
	TrackingNumber string         `json:"trackingNumber"`
	Courier        string         `json:"courier"`
	Notes          string         `json:"notes"`
}

// ── Interfaces ────────────────────────────────────────────────────────────────

type OrderRepository interface {
	Create(o *Order) error
	FindByID(id uuid.UUID) (*Order, error)
	FindByOrderNo(no string) (*Order, error)
	FindByPaymentRef(ref string) (*Order, error)
	FindByUserID(userID uuid.UUID, page, pageSize int) ([]Order, int64, error)
	Update(o *Order) error
	List(status OrderStatus, search string, page, pageSize int) ([]Order, int64, error)
	CountByStatus(status OrderStatus) (int64, error)
	TotalRevenue() (float64, error)
	RevenueByMonth(months int) ([]MonthlyRevenue, error)

	// Cart
	UpsertCartItem(item *CartItem) error
	GetCartItems(userID uuid.UUID) ([]CartItem, error)
	DeleteCartItem(id uuid.UUID, userID uuid.UUID) error
	ClearCart(userID uuid.UUID) error
	FindCartItem(userID, variantID uuid.UUID) (*CartItem, error)
}

type OrderService interface {
	GetCart(userID uuid.UUID) ([]CartItem, error)
	AddToCart(userID uuid.UUID, req CartItemRequest) ([]CartItem, error)
	UpdateCartItem(userID uuid.UUID, itemID uuid.UUID, qty int) ([]CartItem, error)
	RemoveCartItem(userID uuid.UUID, itemID uuid.UUID) error
	Checkout(userID uuid.UUID, req CheckoutRequest) (*Order, string, error) // order, stripe client secret
	GuestCheckout(req GuestCheckoutRequest) (*Order, error)
	CaptureStripePayment(userID uuid.UUID, orderNo, sessionID string) (*Order, error)
	ConfirmStripePayment(sessionID string) error
	GetOrder(id uuid.UUID, userID uuid.UUID) (*Order, error)
	GetOrderByNo(userID uuid.UUID, orderNo string) (*Order, error)
	ListUserOrders(userID uuid.UUID, page, pageSize int) ([]Order, int64, error)
	ListAllOrders(status OrderStatus, search string, page, pageSize int) ([]Order, int64, error)
	UpdateFulfillment(orderID uuid.UUID, req UpdateFulfillmentRequest) (*Order, error)
	CancelOrder(orderID uuid.UUID, userID uuid.UUID) error
	GetDashboardStats() (*DashboardStats, error)
}

// ── Analytics / Dashboard DTOs ────────────────────────────────────────────────

type MonthlyRevenue struct {
	Month   string  `json:"month"` // "2026-01"
	Revenue float64 `json:"revenue"`
	Orders  int64   `json:"orders"`
}

type DashboardStats struct {
	TotalRevenue     float64          `json:"totalRevenue"`
	TotalOrders      int64            `json:"totalOrders"`
	TotalCustomers   int64            `json:"totalCustomers"`
	TotalDealers     int64            `json:"totalDealers"`
	PendingOrders    int64            `json:"pendingOrders"`
	LowStockProducts int64            `json:"lowStockProducts"`
	MonthlyRevenue   []MonthlyRevenue `json:"monthlyRevenue"`
}
