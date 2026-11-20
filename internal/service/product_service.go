package service

import (
	"fmt"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
)

type productService struct {
	productRepo domain.ProductRepository
	storageSvc  domain.StorageClient
}

func NewProductService(productRepo domain.ProductRepository, storageSvc domain.StorageClient) domain.ProductService {
	return &productService{productRepo: productRepo, storageSvc: storageSvc}
}

func (s *productService) CreateProduct(adminID uuid.UUID, req domain.CreateProductRequest) (*domain.Product, error) {
	status := domain.ProductStatus(req.Status)
	if status == "" {
		status = domain.ProductStatusDraft
	}
	p := &domain.Product{
		Name:        req.Name,
		NameZH:      req.NameZH,
		Description: req.Description,
		Category:    req.Category,
		Brand:       req.Brand,
		Tags:        req.Tags,
		Status:      status,
		IsFeatured:  req.IsFeatured,
		SortOrder:   req.SortOrder,
		CreatedBy:   adminID,
	}
	// Generate SKU if not provided
	p.SKU = generateSKU()
	if err := s.productRepo.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *productService) GetProduct(id uuid.UUID) (*domain.Product, error) {
	return s.productRepo.FindByID(id)
}

func (s *productService) UpdateProduct(id uuid.UUID, req domain.UpdateProductRequest) (*domain.Product, error) {
	p, err := s.productRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.NameZH != nil {
		p.NameZH = *req.NameZH
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if req.Category != nil {
		p.Category = *req.Category
	}
	if req.Brand != nil {
		p.Brand = *req.Brand
	}
	if req.Tags != nil {
		p.Tags = *req.Tags
	}
	if req.Status != nil {
		p.Status = domain.ProductStatus(*req.Status)
	}
	if req.IsFeatured != nil {
		p.IsFeatured = *req.IsFeatured
	}
	if req.SortOrder != nil {
		p.SortOrder = *req.SortOrder
	}
	if err := s.productRepo.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *productService) DeleteProduct(id uuid.UUID) error {
	return s.productRepo.Delete(id)
}

func (s *productService) ListProducts(status domain.ProductStatus, category, search string, page, pageSize int) ([]domain.Product, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.productRepo.List(status, category, search, page, pageSize)
}

func (s *productService) ListPublishedProducts(category, search string, page, pageSize int) ([]domain.Product, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.productRepo.ListPublished(category, search, page, pageSize)
}

func (s *productService) ListFeaturedProducts() ([]domain.Product, error) {
	return s.productRepo.ListFeatured()
}

func (s *productService) UploadProductImage(productID uuid.UUID, variantID *uuid.UUID, data []byte, contentType string) (string, error) {
	if s.storageSvc == nil {
		return "", fmt.Errorf("storage not configured")
	}
	key := fmt.Sprintf("products/%s/main", productID)
	if variantID != nil {
		key = fmt.Sprintf("products/%s/variants/%s", productID, *variantID)
	}
	url, err := s.storageSvc.UploadFile("asianmum", key, data, contentType)
	if err != nil {
		return "", err
	}
	// Persist URL
	if variantID != nil {
		v, err := s.productRepo.FindVariantByID(*variantID)
		if err == nil {
			v.ImageURL = url
			s.productRepo.UpdateVariant(v)
		}
	} else {
		p, err := s.productRepo.FindByID(productID)
		if err == nil {
			p.ImageURL = url
			s.productRepo.Update(p)
		}
	}
	return url, nil
}

func (s *productService) CreateVariant(productID uuid.UUID, req domain.CreateVariantRequest) (*domain.ProductVariant, error) {
	v := &domain.ProductVariant{
		ProductID:     productID,
		Name:          req.Name,
		SKU:           req.SKU,
		Attributes:    req.Attributes,
		RetailPrice:   req.RetailPrice,
		DealerPrice:   req.DealerPrice,
		BV:            req.BV,
		PV:            req.PV,
		StockOnHand:   req.StockOnHand,
		LowStockAlert: req.LowStockAlert,
		Weight:        req.Weight,
		IsActive:      req.IsActive,
		SortOrder:     req.SortOrder,
	}
	if v.LowStockAlert == 0 {
		v.LowStockAlert = 10
	}
	return v, s.productRepo.CreateVariant(v)
}

func (s *productService) UpdateVariant(variantID uuid.UUID, req domain.UpdateVariantRequest) (*domain.ProductVariant, error) {
	v, err := s.productRepo.FindVariantByID(variantID)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		v.Name = *req.Name
	}
	if req.Attributes != nil {
		v.Attributes = *req.Attributes
	}
	if req.RetailPrice != nil {
		v.RetailPrice = *req.RetailPrice
	}
	if req.DealerPrice != nil {
		v.DealerPrice = *req.DealerPrice
	}
	if req.BV != nil {
		v.BV = *req.BV
	}
	if req.PV != nil {
		v.PV = *req.PV
	}
	if req.LowStockAlert != nil {
		v.LowStockAlert = *req.LowStockAlert
	}
	if req.Weight != nil {
		v.Weight = *req.Weight
	}
	if req.IsActive != nil {
		v.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		v.SortOrder = *req.SortOrder
	}
	return v, s.productRepo.UpdateVariant(v)
}

func (s *productService) DeleteVariant(variantID uuid.UUID) error {
	return s.productRepo.DeleteVariant(variantID)
}

func (s *productService) CreateReview(userID uuid.UUID, req domain.CreateReviewRequest) (*domain.ProductReview, error) {
	// Prevent duplicate reviews
	already, _ := s.productRepo.HasReviewed(userID, req.OrderID, req.ProductID)
	if already {
		return nil, fmt.Errorf("you have already reviewed this product for this order")
	}
	r := &domain.ProductReview{
		ProductID: req.ProductID,
		UserID:    userID,
		OrderID:   req.OrderID,
		Rating:    req.Rating,
		Title:     req.Title,
		Body:      req.Body,
		IsVisible: true,
	}
	return r, s.productRepo.CreateReview(r)
}

func (s *productService) ListReviews(productID uuid.UUID, page, pageSize int) ([]domain.ProductReview, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}
	return s.productRepo.ListReviews(productID, page, pageSize)
}

// generateSKU creates a short unique product code.
func generateSKU() string {
	b := make([]byte, 3)
	fmt.Sprintf("%x", b)
	id := uuid.New().String()
	return fmt.Sprintf("HM-%s", id[:8])
}
