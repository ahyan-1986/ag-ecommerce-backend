package service

import (
	"errors"
	"fmt"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo   domain.UserRepository
	storageSvc domain.StorageClient
}

func NewUserService(userRepo domain.UserRepository, storageSvc domain.StorageClient) domain.UserService {
	return &userService{userRepo: userRepo, storageSvc: storageSvc}
}

func (s *userService) GetProfile(userID uuid.UUID) (*domain.UserPublicProfile, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	p := toPublicProfile(user)
	return &p, nil
}

func (s *userService) UpdateProfile(userID uuid.UUID, req domain.UpdateProfileRequest) (*domain.UserPublicProfile, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Phone != nil {
		// Check uniqueness if phone changed
		if *req.Phone != user.Phone {
			existing, err := s.userRepo.FindByPhone(*req.Phone)
			if err == nil && existing.ID != userID {
				return nil, errors.New("phone number already used by another account")
			}
		}
		user.Phone = *req.Phone
	}
	if req.DateOfBirth != nil {
		user.DateOfBirth = req.DateOfBirth
	}
	if req.Gender != nil {
		user.Gender = *req.Gender
	}
	if req.AddressLine1 != nil {
		user.AddressLine1 = *req.AddressLine1
	}
	if req.AddressLine2 != nil {
		user.AddressLine2 = *req.AddressLine2
	}
	if req.City != nil {
		user.City = *req.City
	}
	if req.State != nil {
		user.State = *req.State
	}
	if req.Postcode != nil {
		user.Postcode = *req.Postcode
	}
	if req.CompanyName != nil && user.IsDealer() {
		user.CompanyName = *req.CompanyName
	}
	if req.BankName != nil {
		user.BankName = *req.BankName
	}
	if req.BankAccountNo != nil {
		user.BankAccountNo = *req.BankAccountNo
	}
	if req.PayeeName != nil {
		user.PayeeName = *req.PayeeName
	}
	if req.PayeeIDNo != nil {
		user.PayeeIDNo = *req.PayeeIDNo
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}
	p := toPublicProfile(user)
	return &p, nil
}

func (s *userService) ChangePassword(userID uuid.UUID, req domain.ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hash)
	return s.userRepo.Update(user)
}

func (s *userService) UploadAvatar(userID uuid.UUID, data []byte, contentType string) (string, error) {
	if s.storageSvc == nil {
		return "", errors.New("storage not configured")
	}
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", err
	}
	key := fmt.Sprintf("avatars/%s", userID.String())
	url, err := s.storageSvc.UploadFile("asianmum", key, data, contentType)
	if err != nil {
		return "", err
	}
	user.AvatarURL = url
	if err := s.userRepo.Update(user); err != nil {
		return "", err
	}
	return url, nil
}

func (s *userService) ListUsers(role domain.UserRole, search string, page, pageSize int) ([]domain.UserPublicProfile, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	users, total, err := s.userRepo.List(role, search, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	profiles := make([]domain.UserPublicProfile, len(users))
	for i, u := range users {
		profiles[i] = toPublicProfile(&u)
	}
	return profiles, total, nil
}

func (s *userService) GetUser(id uuid.UUID) (*domain.UserPublicProfile, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	p := toPublicProfile(user)
	return &p, nil
}

func (s *userService) SetActive(id uuid.UUID, active bool) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}
	user.IsActive = active
	return s.userRepo.Update(user)
}

func (s *userService) ApproveDealerApplication(id uuid.UUID) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}
	if !user.IsDealer() {
		return errors.New("user is not a dealer")
	}
	user.DealerApproved = true
	return s.userRepo.Update(user)
}

func (s *userService) GetCRMSummary() (*domain.CRMSummary, error) {
	return s.userRepo.GetCRMSummary()
}

func (s *userService) GetUserCRMStats(userID uuid.UUID) (*domain.CRMUserStats, error) {
	return s.userRepo.GetUserCRMStats(userID)
}

