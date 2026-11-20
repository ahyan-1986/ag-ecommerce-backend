package service

import (
	"errors"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
)

type addressBookService struct {
	repo domain.AddressBookRepository
}

func NewAddressBookService(repo domain.AddressBookRepository) domain.AddressBookService {
	return &addressBookService{repo: repo}
}

func (s *addressBookService) List(userID uuid.UUID) ([]domain.AddressBook, error) {
	return s.repo.List(userID)
}

func (s *addressBookService) Create(req *domain.CreateAddressBookRequest, userID uuid.UUID) (*domain.AddressBook, error) {
	if req.IsDefault {
		if err := s.repo.ClearDefaults(userID); err != nil {
			return nil, err
		}
	}
	country := req.Country
	if country == "" {
		country = "Malaysia"
	}
	addr := &domain.AddressBook{
		UserID:        userID,
		Label:         req.Label,
		RecipientName: req.RecipientName,
		Phone:         req.Phone,
		AddressLine1:  req.AddressLine1,
		AddressLine2:  req.AddressLine2,
		City:          req.City,
		State:         req.State,
		Postcode:      req.Postcode,
		Country:       country,
		IsDefault:     req.IsDefault,
	}
	if err := s.repo.Create(addr); err != nil {
		return nil, err
	}
	return addr, nil
}

func (s *addressBookService) Update(id uuid.UUID, req *domain.UpdateAddressBookRequest, userID uuid.UUID) (*domain.AddressBook, error) {
	addr, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, errors.New("address not found")
	}
	if req.Label != nil {
		addr.Label = *req.Label
	}
	if req.RecipientName != nil {
		addr.RecipientName = *req.RecipientName
	}
	if req.Phone != nil {
		addr.Phone = *req.Phone
	}
	if req.AddressLine1 != nil {
		addr.AddressLine1 = *req.AddressLine1
	}
	if req.AddressLine2 != nil {
		addr.AddressLine2 = *req.AddressLine2
	}
	if req.City != nil {
		addr.City = *req.City
	}
	if req.State != nil {
		addr.State = *req.State
	}
	if req.Postcode != nil {
		addr.Postcode = *req.Postcode
	}
	if req.Country != nil {
		addr.Country = *req.Country
	}
	if req.IsDefault != nil {
		if *req.IsDefault {
			if err := s.repo.ClearDefaults(userID); err != nil {
				return nil, err
			}
		}
		addr.IsDefault = *req.IsDefault
	}
	if err := s.repo.Save(addr); err != nil {
		return nil, err
	}
	return addr, nil
}

func (s *addressBookService) Delete(id, userID uuid.UUID) error {
	_, err := s.repo.FindByID(id, userID)
	if err != nil {
		return errors.New("address not found")
	}
	return s.repo.Delete(id, userID)
}

func (s *addressBookService) SetDefault(id, userID uuid.UUID) (*domain.AddressBook, error) {
	addr, err := s.repo.FindByID(id, userID)
	if err != nil {
		return nil, errors.New("address not found")
	}
	if err := s.repo.ClearDefaults(userID); err != nil {
		return nil, err
	}
	addr.IsDefault = true
	if err := s.repo.Save(addr); err != nil {
		return nil, err
	}
	return addr, nil
}
