package repository

import (
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BannerRepository interface {
	List(activeOnly bool) ([]domain.Banner, error)
	FindByID(id uuid.UUID) (*domain.Banner, error)
	Create(banner *domain.Banner) error
	Update(banner *domain.Banner) error
	Delete(id uuid.UUID) error
}

type bannerRepository struct {
	db *gorm.DB
}

func NewBannerRepository(db *gorm.DB) BannerRepository {
	return &bannerRepository{db: db}
}

func (r *bannerRepository) List(activeOnly bool) ([]domain.Banner, error) {
	var banners []domain.Banner
	query := r.db.Model(&domain.Banner{})
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	err := query.Order("sort_order asc, created_at desc").Find(&banners).Error
	return banners, err
}

func (r *bannerRepository) FindByID(id uuid.UUID) (*domain.Banner, error) {
	var banner domain.Banner
	err := r.db.First(&banner, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &banner, nil
}

func (r *bannerRepository) Create(banner *domain.Banner) error {
	return r.db.Create(banner).Error
}

func (r *bannerRepository) Update(banner *domain.Banner) error {
	return r.db.Save(banner).Error
}

func (r *bannerRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Banner{}, "id = ?", id).Error
}
