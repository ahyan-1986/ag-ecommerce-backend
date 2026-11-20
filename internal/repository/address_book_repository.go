package repository

import (
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type addressBookRepository struct {
	db *gorm.DB
}

func NewAddressBookRepository(db *gorm.DB) domain.AddressBookRepository {
	return &addressBookRepository{db: db}
}

func (r *addressBookRepository) List(userID uuid.UUID) ([]domain.AddressBook, error) {
	var items []domain.AddressBook
	err := r.db.
		Where("user_id = ?", userID).
		Order("is_default DESC, created_at ASC").
		Find(&items).Error
	return items, err
}

func (r *addressBookRepository) FindByID(id, userID uuid.UUID) (*domain.AddressBook, error) {
	var item domain.AddressBook
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *addressBookRepository) Create(addr *domain.AddressBook) error {
	return r.db.Create(addr).Error
}

func (r *addressBookRepository) Save(addr *domain.AddressBook) error {
	return r.db.Save(addr).Error
}

func (r *addressBookRepository) Delete(id, userID uuid.UUID) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&domain.AddressBook{}).Error
}

func (r *addressBookRepository) ClearDefaults(userID uuid.UUID) error {
	return r.db.Model(&domain.AddressBook{}).
		Where("user_id = ? AND is_default = true", userID).
		Update("is_default", false).Error
}
