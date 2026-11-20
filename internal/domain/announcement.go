package domain

import (
	"time"

	"github.com/google/uuid"
)

type Announcement struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title     string    `json:"title" gorm:"not null"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	IsActive  bool      `json:"isActive" gorm:"default:true"`
	Priority       int       `json:"priority" gorm:"default:0"` // Higher number means more important
	AttachmentURL  string    `json:"attachmentUrl" gorm:"type:text"`
	AttachmentName string    `json:"attachmentName" gorm:"type:varchar(255)"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
