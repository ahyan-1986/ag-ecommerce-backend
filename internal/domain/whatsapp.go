package domain

import (
	"time"

	"github.com/google/uuid"
)

// WhatsAppConfig holds Meta Business API credentials (one record per system).
type WhatsAppConfig struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PhoneNumberID   string    `gorm:"type:varchar(100);not null" json:"phoneNumberId"`
	BusinessID      string    `gorm:"type:varchar(100);default:''" json:"businessId"`
	AccessToken     string    `gorm:"type:text;not null" json:"-"`
	DisplayPhone    string    `gorm:"type:varchar(30);default:''" json:"displayPhone"`
	WebhookVerifyTk string    `gorm:"type:varchar(200);default:''" json:"-"`
	Enabled         bool      `gorm:"not null;default:false" json:"enabled"`
	UpdatedBy       uuid.UUID `gorm:"type:uuid" json:"updatedBy"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (WhatsAppConfig) TableName() string { return "whatsapp_configs" }

// WhatsAppTemplate is a Meta-approved message template.
type WhatsAppTemplate struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MetaTemplateID string    `gorm:"type:varchar(100);default:''" json:"metaTemplateId"`
	Name           string    `gorm:"type:varchar(200);not null" json:"name"`
	Language       string    `gorm:"type:varchar(10);not null;default:'en'" json:"language"`
	Category       string    `gorm:"type:varchar(50);not null;default:'MARKETING'" json:"category"`
	Status         string    `gorm:"type:varchar(30);not null;default:'PENDING'" json:"status"`
	HeaderText     string    `gorm:"type:text;default:''" json:"headerText"`
	BodyText       string    `gorm:"type:text;not null" json:"bodyText"`
	FooterText     string    `gorm:"type:varchar(200);default:''" json:"footerText"`
	ButtonsJSON    string    `gorm:"type:text;default:'[]'" json:"buttonsJson"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (WhatsAppTemplate) TableName() string { return "whatsapp_templates" }

// WhatsAppBlast is a bulk WhatsApp marketing campaign.
type BlastStatus string

const (
	BlastPending    BlastStatus = "pending"
	BlastProcessing BlastStatus = "processing"
	BlastCompleted  BlastStatus = "completed"
	BlastFailed     BlastStatus = "failed"
)

type WhatsAppBlast struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name         string    `gorm:"type:varchar(200);not null" json:"name"`
	TemplateID   uuid.UUID `gorm:"type:uuid;not null" json:"templateId"`
	TemplateName string    `gorm:"type:varchar(200);default:''" json:"templateName"`
	// Target audience: all, customers, dealers, tag
	TargetAudience string      `gorm:"type:varchar(50);not null;default:'all'" json:"targetAudience"`
	TemplateParams string      `gorm:"type:text;default:'{}'" json:"templateParams"` // JSON variable substitutions
	Status         BlastStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	TotalCount     int         `gorm:"default:0" json:"totalCount"`
	SentCount      int         `gorm:"default:0" json:"sentCount"`
	FailedCount    int         `gorm:"default:0" json:"failedCount"`
	ScheduledAt    *time.Time  `gorm:"type:timestamptz" json:"scheduledAt,omitempty"`
	StartedAt      *time.Time  `gorm:"type:timestamptz" json:"startedAt,omitempty"`
	CompletedAt    *time.Time  `gorm:"type:timestamptz" json:"completedAt,omitempty"`
	CreatedBy      uuid.UUID   `gorm:"type:uuid;not null" json:"createdBy"`
	CreatedAt      time.Time   `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time   `gorm:"autoUpdateTime" json:"updatedAt"`

	Recipients []WhatsAppBlastRecipient `gorm:"foreignKey:BlastID" json:"recipients,omitempty"`
}

func (WhatsAppBlast) TableName() string { return "whatsapp_blasts" }

type WhatsAppBlastRecipient struct {
	ID       uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BlastID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"blastId"`
	UserID   uuid.UUID  `gorm:"type:uuid;not null" json:"userId"`
	Phone    string     `gorm:"type:varchar(30);not null" json:"phone"`
	Status   string     `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	ErrorMsg string     `gorm:"type:varchar(500);default:''" json:"errorMsg"`
	SentAt   *time.Time `gorm:"type:timestamptz" json:"sentAt,omitempty"`
}

func (WhatsAppBlastRecipient) TableName() string { return "whatsapp_blast_recipients" }

// ── Request / Response DTOs ───────────────────────────────────────────────────

type SaveWhatsAppConfigRequest struct {
	PhoneNumberID   string `json:"phoneNumberId" binding:"required"`
	BusinessID      string `json:"businessId"`
	AccessToken     string `json:"accessToken" binding:"required"`
	DisplayPhone    string `json:"displayPhone"`
	WebhookVerifyTk string `json:"webhookVerifyToken"`
	Enabled         bool   `json:"enabled"`
}

type CreateTemplateRequest struct {
	Name        string `json:"name" binding:"required"`
	Language    string `json:"language"`
	Category    string `json:"category"`
	HeaderText  string `json:"headerText"`
	BodyText    string `json:"bodyText" binding:"required"`
	FooterText  string `json:"footerText"`
	ButtonsJSON string `json:"buttonsJson"`
}

type CreateBlastRequest struct {
	Name           string     `json:"name" binding:"required"`
	TemplateID     uuid.UUID  `json:"templateId" binding:"required"`
	TargetAudience string     `json:"targetAudience"` // all, customers, dealers
	TemplateParams string     `json:"templateParams"`
	ScheduledAt    *time.Time `json:"scheduledAt"`
}

// ── Interfaces ────────────────────────────────────────────────────────────────

type WhatsAppRepository interface {
	GetConfig() (*WhatsAppConfig, error)
	SaveConfig(c *WhatsAppConfig) error
	CreateTemplate(t *WhatsAppTemplate) error
	FindTemplateByID(id uuid.UUID) (*WhatsAppTemplate, error)
	ListTemplates() ([]WhatsAppTemplate, error)
	DeleteTemplate(id uuid.UUID) error
	CreateBlast(b *WhatsAppBlast) error
	FindBlastByID(id uuid.UUID) (*WhatsAppBlast, error)
	UpdateBlast(b *WhatsAppBlast) error
	ListBlasts(page, pageSize int) ([]WhatsAppBlast, int64, error)
	CreateRecipient(r *WhatsAppBlastRecipient) error
	UpdateRecipient(r *WhatsAppBlastRecipient) error
	ListRecipients(blastID uuid.UUID) ([]WhatsAppBlastRecipient, error)
}

type WhatsAppService interface {
	GetConfig() (*WhatsAppConfig, error)
	SaveConfig(adminID uuid.UUID, req SaveWhatsAppConfigRequest) (*WhatsAppConfig, error)
	CreateTemplate(req CreateTemplateRequest) (*WhatsAppTemplate, error)
	ListTemplates() ([]WhatsAppTemplate, error)
	DeleteTemplate(id uuid.UUID) error
	CreateBlast(adminID uuid.UUID, req CreateBlastRequest) (*WhatsAppBlast, error)
	SendBlast(blastID uuid.UUID) error
	GetBlast(id uuid.UUID) (*WhatsAppBlast, error)
	ListBlasts(page, pageSize int) ([]WhatsAppBlast, int64, error)
}

// WhatsAppClient is the interface to the Meta Cloud API.
type WhatsAppClient interface {
	SendTemplate(phoneNumberID, accessToken, toPhone, templateName, language string, params []string) error
}
