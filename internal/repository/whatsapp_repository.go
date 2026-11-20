package repository

import (
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type whatsappRepository struct{ db *gorm.DB }

func NewWhatsAppRepository(db *gorm.DB) domain.WhatsAppRepository {
	return &whatsappRepository{db: db}
}

func (r *whatsappRepository) GetConfig() (*domain.WhatsAppConfig, error) {
	var c domain.WhatsAppConfig
	err := r.db.First(&c).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &c, err
}

func (r *whatsappRepository) SaveConfig(c *domain.WhatsAppConfig) error {
	existing, _ := r.GetConfig()
	if existing == nil {
		return r.db.Create(c).Error
	}
	c.ID = existing.ID
	return r.db.Save(c).Error
}

func (r *whatsappRepository) CreateTemplate(t *domain.WhatsAppTemplate) error {
	return r.db.Create(t).Error
}

func (r *whatsappRepository) FindTemplateByID(id uuid.UUID) (*domain.WhatsAppTemplate, error) {
	var t domain.WhatsAppTemplate
	err := r.db.First(&t, "id = ?", id).Error
	return &t, err
}

func (r *whatsappRepository) ListTemplates() ([]domain.WhatsAppTemplate, error) {
	var items []domain.WhatsAppTemplate
	err := r.db.Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *whatsappRepository) DeleteTemplate(id uuid.UUID) error {
	return r.db.Delete(&domain.WhatsAppTemplate{}, "id = ?", id).Error
}

func (r *whatsappRepository) CreateBlast(b *domain.WhatsAppBlast) error {
	return r.db.Create(b).Error
}

func (r *whatsappRepository) FindBlastByID(id uuid.UUID) (*domain.WhatsAppBlast, error) {
	var b domain.WhatsAppBlast
	err := r.db.Preload("Recipients").First(&b, "id = ?", id).Error
	return &b, err
}

func (r *whatsappRepository) UpdateBlast(b *domain.WhatsAppBlast) error {
	return r.db.Save(b).Error
}

func (r *whatsappRepository) ListBlasts(page, pageSize int) ([]domain.WhatsAppBlast, int64, error) {
	q := r.db.Model(&domain.WhatsAppBlast{})
	var total int64
	q.Count(&total)
	var items []domain.WhatsAppBlast
	offset := (page - 1) * pageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *whatsappRepository) CreateRecipient(rc *domain.WhatsAppBlastRecipient) error {
	return r.db.Create(rc).Error
}

func (r *whatsappRepository) UpdateRecipient(rc *domain.WhatsAppBlastRecipient) error {
	return r.db.Save(rc).Error
}

func (r *whatsappRepository) ListRecipients(blastID uuid.UUID) ([]domain.WhatsAppBlastRecipient, error) {
	var items []domain.WhatsAppBlastRecipient
	err := r.db.Where("blast_id = ?", blastID).Find(&items).Error
	return items, err
}
