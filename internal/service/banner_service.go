package service

import (
	"fmt"
	"time"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/repository"
	"github.com/google/uuid"
)

type BannerService interface {
	ListBanners(activeOnly bool) ([]domain.Banner, error)
	CreateBanner(banner *domain.Banner) error
	UpdateBanner(id uuid.UUID, req *domain.Banner) error
	DeleteBanner(id uuid.UUID) error
	UploadBannerImage(id uuid.UUID, data []byte, contentType string) (string, error)
}

type bannerService struct {
	repo    repository.BannerRepository
	storage domain.StorageClient
}

func NewBannerService(repo repository.BannerRepository, storageClient domain.StorageClient) BannerService {
	return &bannerService{repo: repo, storage: storageClient}
}

func (s *bannerService) ListBanners(activeOnly bool) ([]domain.Banner, error) {
	return s.repo.List(activeOnly)
}

func (s *bannerService) CreateBanner(banner *domain.Banner) error {
	if banner.ID == uuid.Nil {
		banner.ID = uuid.New()
	}
	return s.repo.Create(banner)
}

func (s *bannerService) UpdateBanner(id uuid.UUID, req *domain.Banner) error {
	banner, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	banner.Title = req.Title
	banner.Subtitle = req.Subtitle
	banner.LinkURL = req.LinkURL
	banner.IsActive = req.IsActive
	banner.SortOrder = req.SortOrder
	return s.repo.Update(banner)
}

func (s *bannerService) DeleteBanner(id uuid.UUID) error {
	// Optional: delete image from S3? We could just leave it or delete it.
	// For simplicity, we just delete the DB record.
	return s.repo.Delete(id)
}

func (s *bannerService) UploadBannerImage(id uuid.UUID, data []byte, contentType string) (string, error) {
	banner, err := s.repo.FindByID(id)
	if err != nil {
		return "", err
	}

	if s.storage == nil {
		return "", fmt.Errorf("storage service not configured")
	}

	ext := ".png"
	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/webp":
		ext = ".webp"
	}
	key := fmt.Sprintf("banners/%s_%d%s", id.String(), time.Now().Unix(), ext)

	url, err := s.storage.UploadFile("asianmum", key, data, contentType)
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
	}

	banner.ImageURL = url
	if err := s.repo.Update(banner); err != nil {
		return "", err
	}

	return url, nil
}
