package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
)

type whatsappService struct {
	waRepo   domain.WhatsAppRepository
	userRepo domain.UserRepository
	client   domain.WhatsAppClient
}

func NewWhatsAppService(waRepo domain.WhatsAppRepository, userRepo domain.UserRepository, client domain.WhatsAppClient) domain.WhatsAppService {
	return &whatsappService{waRepo: waRepo, userRepo: userRepo, client: client}
}

func (s *whatsappService) GetConfig() (*domain.WhatsAppConfig, error) {
	return s.waRepo.GetConfig()
}

func (s *whatsappService) SaveConfig(adminID uuid.UUID, req domain.SaveWhatsAppConfigRequest) (*domain.WhatsAppConfig, error) {
	cfg := &domain.WhatsAppConfig{
		PhoneNumberID:   req.PhoneNumberID,
		BusinessID:      req.BusinessID,
		AccessToken:     req.AccessToken,
		DisplayPhone:    req.DisplayPhone,
		WebhookVerifyTk: req.WebhookVerifyTk,
		Enabled:         req.Enabled,
		UpdatedBy:       adminID,
	}
	return cfg, s.waRepo.SaveConfig(cfg)
}

func (s *whatsappService) CreateTemplate(req domain.CreateTemplateRequest) (*domain.WhatsAppTemplate, error) {
	t := &domain.WhatsAppTemplate{
		Name:        req.Name,
		Language:    req.Language,
		Category:    req.Category,
		HeaderText:  req.HeaderText,
		BodyText:    req.BodyText,
		FooterText:  req.FooterText,
		ButtonsJSON: req.ButtonsJSON,
		Status:      "PENDING",
	}
	if t.Language == "" {
		t.Language = "en"
	}
	if t.Category == "" {
		t.Category = "MARKETING"
	}
	return t, s.waRepo.CreateTemplate(t)
}

func (s *whatsappService) ListTemplates() ([]domain.WhatsAppTemplate, error) {
	return s.waRepo.ListTemplates()
}

func (s *whatsappService) DeleteTemplate(id uuid.UUID) error {
	return s.waRepo.DeleteTemplate(id)
}

func (s *whatsappService) CreateBlast(adminID uuid.UUID, req domain.CreateBlastRequest) (*domain.WhatsAppBlast, error) {
	tmpl, err := s.waRepo.FindTemplateByID(req.TemplateID)
	if err != nil {
		return nil, errors.New("template not found")
	}
	audience := req.TargetAudience
	if audience == "" {
		audience = "all"
	}
	b := &domain.WhatsAppBlast{
		Name:           req.Name,
		TemplateID:     req.TemplateID,
		TemplateName:   tmpl.Name,
		TargetAudience: audience,
		TemplateParams: req.TemplateParams,
		Status:         domain.BlastPending,
		ScheduledAt:    req.ScheduledAt,
		CreatedBy:      adminID,
	}
	return b, s.waRepo.CreateBlast(b)
}

func (s *whatsappService) SendBlast(blastID uuid.UUID) error {
	blast, err := s.waRepo.FindBlastByID(blastID)
	if err != nil {
		return errors.New("blast not found")
	}
	if blast.Status == domain.BlastCompleted || blast.Status == domain.BlastProcessing {
		return errors.New("blast already sent or in progress")
	}

	cfg, err := s.waRepo.GetConfig()
	if err != nil || cfg == nil || !cfg.Enabled {
		return errors.New("WhatsApp not configured or disabled")
	}

	tmpl, err := s.waRepo.FindTemplateByID(blast.TemplateID)
	if err != nil {
		return errors.New("template not found")
	}

	// Determine target users
	var role domain.UserRole
	switch blast.TargetAudience {
	case "customers":
		role = domain.RoleCustomer
	case "dealers":
		role = domain.RoleDealer
	}
	users, _, err := s.userRepo.List(role, "", 1, 10000)
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}

	// Create recipients
	for _, u := range users {
		if u.Phone == "" {
			continue
		}
		rc := &domain.WhatsAppBlastRecipient{
			BlastID: blastID,
			UserID:  u.ID,
			Phone:   u.Phone,
			Status:  "pending",
		}
		s.waRepo.CreateRecipient(rc)
	}

	// Mark blast as processing
	now := time.Now()
	blast.Status = domain.BlastProcessing
	blast.StartedAt = &now
	blast.TotalCount = len(users)
	s.waRepo.UpdateBlast(blast)

	// Send in goroutine
	go func() {
		recipients, _ := s.waRepo.ListRecipients(blastID)
		for _, rc := range recipients {
			if rc.Phone == "" {
				continue
			}
			err := s.client.SendTemplate(
				cfg.PhoneNumberID,
				cfg.AccessToken,
				rc.Phone,
				tmpl.Name,
				tmpl.Language,
				[]string{},
			)
			now := time.Now()
			if err != nil {
				rc.Status = "failed"
				rc.ErrorMsg = err.Error()
				blast.FailedCount++
			} else {
				rc.Status = "sent"
				rc.SentAt = &now
				blast.SentCount++
			}
			s.waRepo.UpdateRecipient(&rc)
		}
		blast.Status = domain.BlastCompleted
		completedAt := time.Now()
		blast.CompletedAt = &completedAt
		s.waRepo.UpdateBlast(blast)
	}()

	return nil
}

func (s *whatsappService) GetBlast(id uuid.UUID) (*domain.WhatsAppBlast, error) {
	return s.waRepo.FindBlastByID(id)
}

func (s *whatsappService) ListBlasts(page, pageSize int) ([]domain.WhatsAppBlast, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}
	return s.waRepo.ListBlasts(page, pageSize)
}
