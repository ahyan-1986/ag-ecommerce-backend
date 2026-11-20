package domain

import (
	"time"

	"github.com/google/uuid"
)

type Banner struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ImageURL  string    `gorm:"type:varchar(255);not null" json:"imageUrl"`
	Title     string    `gorm:"type:varchar(255)" json:"title"`
	Subtitle  string    `gorm:"type:varchar(255)" json:"subtitle"`
	LinkURL   string    `gorm:"type:varchar(255)" json:"linkUrl"`
	IsActive  bool      `gorm:"default:true" json:"isActive"`
	SortOrder int       `gorm:"default:0" json:"sortOrder"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
