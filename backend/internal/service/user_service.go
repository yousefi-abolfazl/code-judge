package service

import (
	"errors"

	"github.com/yousefi-abolfazl/code-judge/backend/internal/models"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/repository"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

type UserProfile struct {
	User  *models.User          `json:"user"`
	Stats *repository.UserStats `json:"stats"`
}

func (s *UserService) GetUserProfile(userID uint) (*UserProfile, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // User not found
		}
		return nil, err
	}

	stats, err := s.userRepo.GetUserStats(userID)
	if err != nil {
		return nil, err
	}

	return &UserProfile{User: user, Stats: stats}, nil
}

func (s *UserService) UpdateUserRole(userID uint, newRole string) error {
	if newRole != models.RoleAdmin && newRole != models.RoleUser {
		return errors.New("invalid role specified")
	}

	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	return s.userRepo.UpdateRole(userID, newRole)
}
