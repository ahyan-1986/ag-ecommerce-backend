package repository

import (
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AnnouncementRepository interface {
	Create(a *domain.Announcement) error
	Update(a *domain.Announcement) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*domain.Announcement, error)
	ListAll() ([]domain.Announcement, error)
	ListActive() ([]domain.Announcement, error)
}

type announcementRepository struct {
	db *gorm.DB
}

func NewAnnouncementRepository(db *gorm.DB) AnnouncementRepository {
	return &announcementRepository{db}
}

func (r *announcementRepository) Create(a *domain.Announcement) error {
	return r.db.Create(a).Error
}

func (r *announcementRepository) Update(a *domain.Announcement) error {
	return r.db.Save(a).Error
}

func (r *announcementRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Announcement{}, "id = ?", id).Error
}

func (r *announcementRepository) FindByID(id uuid.UUID) (*domain.Announcement, error) {
	var a domain.Announcement
	err := r.db.First(&a, "id = ?", id).Error
	return &a, err
}

func (r *announcementRepository) ListAll() ([]domain.Announcement, error) {
	var as []domain.Announcement
	err := r.db.Order("priority DESC, created_at DESC").Find(&as).Error
	return as, err
}

func (r *announcementRepository) ListActive() ([]domain.Announcement, error) {
	var as []domain.Announcement
	err := r.db.Where("is_active = ?", true).Order("priority DESC, created_at DESC").Find(&as).Error
	return as, err
}
