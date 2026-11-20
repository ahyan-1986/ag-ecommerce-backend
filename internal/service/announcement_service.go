package service

import (
	"fmt"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/repository"
	"github.com/google/uuid"
)

type AnnouncementService interface {
	CreateAnnouncement(title, content string, isActive bool, priority int) (*domain.Announcement, error)
	UpdateAnnouncement(id uuid.UUID, title, content string, isActive bool, priority int) error
	DeleteAnnouncement(id uuid.UUID) error
	GetAdminAnnouncements() ([]domain.Announcement, error)
	GetActiveAnnouncements() ([]domain.Announcement, error)
	UploadAttachment(id uuid.UUID, data []byte, contentType, filename string) (string, error)
}

type announcementService struct {
	repo    repository.AnnouncementRepository
	storage domain.StorageClient
}

func NewAnnouncementService(repo repository.AnnouncementRepository, storage domain.StorageClient) AnnouncementService {
	return &announcementService{repo, storage}
}

func (s *announcementService) CreateAnnouncement(title, content string, isActive bool, priority int) (*domain.Announcement, error) {
	a := &domain.Announcement{
		Title:    title,
		Content:  content,
		IsActive: isActive,
		Priority: priority,
	}
	if err := s.repo.Create(a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *announcementService) UpdateAnnouncement(id uuid.UUID, title, content string, isActive bool, priority int) error {
	a, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	a.Title = title
	a.Content = content
	a.IsActive = isActive
	a.Priority = priority

	return s.repo.Update(a)
}

func (s *announcementService) DeleteAnnouncement(id uuid.UUID) error {
	return s.repo.Delete(id)
}

func (s *announcementService) GetAdminAnnouncements() ([]domain.Announcement, error) {
	return s.repo.ListAll()
}

func (s *announcementService) GetActiveAnnouncements() ([]domain.Announcement, error) {
	return s.repo.ListActive()
}

func (s *announcementService) UploadAttachment(id uuid.UUID, data []byte, contentType, filename string) (string, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		return "", err
	}

	if s.storage == nil {
		return "", fmt.Errorf("storage service not configured")
	}

	key := fmt.Sprintf("announcements/%s_%s", id.String(), filename)
	url, err := s.storage.UploadFile("asianmum", key, data, contentType)
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
	}

	a.AttachmentURL = url
	a.AttachmentName = filename
	if err := s.repo.Update(a); err != nil {
		return "", err
	}

	return url, nil
}
