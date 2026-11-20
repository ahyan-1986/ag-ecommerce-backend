package domain

import (
	"time"

	"github.com/google/uuid"
)

// AddressBook stores saved delivery addresses per user.
type AddressBook struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	Label         string    `gorm:"type:varchar(100);default:''" json:"label"` // e.g. "Home", "Office"
	RecipientName string    `gorm:"type:varchar(200);not null" json:"recipientName"`
	Phone         string    `gorm:"type:varchar(50);not null" json:"phone"`
	AddressLine1  string    `gorm:"type:varchar(255);not null" json:"addressLine1"`
	AddressLine2  string    `gorm:"type:varchar(255);default:''" json:"addressLine2"`
	City          string    `gorm:"type:varchar(100);not null" json:"city"`
	State         string    `gorm:"type:varchar(100);default:''" json:"state"`
	Postcode      string    `gorm:"type:varchar(20);default:''" json:"postcode"`
	Country       string    `gorm:"type:varchar(100);not null;default:'Malaysia'" json:"country"`
	IsDefault     bool      `gorm:"not null;default:false" json:"isDefault"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (AddressBook) TableName() string { return "address_books" }

// ── DTOs ─────────────────────────────────────────────────────────────────────

type CreateAddressBookRequest struct {
	Label         string `json:"label"`
	RecipientName string `json:"recipientName" binding:"required"`
	Phone         string `json:"phone" binding:"required"`
	AddressLine1  string `json:"addressLine1" binding:"required"`
	AddressLine2  string `json:"addressLine2"`
	City          string `json:"city" binding:"required"`
	State         string `json:"state"`
	Postcode      string `json:"postcode"`
	Country       string `json:"country"`
	IsDefault     bool   `json:"isDefault"`
}

type UpdateAddressBookRequest struct {
	Label         *string `json:"label"`
	RecipientName *string `json:"recipientName"`
	Phone         *string `json:"phone"`
	AddressLine1  *string `json:"addressLine1"`
	AddressLine2  *string `json:"addressLine2"`
	City          *string `json:"city"`
	State         *string `json:"state"`
	Postcode      *string `json:"postcode"`
	Country       *string `json:"country"`
	IsDefault     *bool   `json:"isDefault"`
}

// ── Repository interface ─────────────────────────────────────────────────────

type AddressBookRepository interface {
	List(userID uuid.UUID) ([]AddressBook, error)
	FindByID(id, userID uuid.UUID) (*AddressBook, error)
	Create(addr *AddressBook) error
	Save(addr *AddressBook) error
	Delete(id, userID uuid.UUID) error
	ClearDefaults(userID uuid.UUID) error
}

// ── Service interface ─────────────────────────────────────────────────────────

type AddressBookService interface {
	List(userID uuid.UUID) ([]AddressBook, error)
	Create(req *CreateAddressBookRequest, userID uuid.UUID) (*AddressBook, error)
	Update(id uuid.UUID, req *UpdateAddressBookRequest, userID uuid.UUID) (*AddressBook, error)
	Delete(id, userID uuid.UUID) error
	SetDefault(id, userID uuid.UUID) (*AddressBook, error)
}
