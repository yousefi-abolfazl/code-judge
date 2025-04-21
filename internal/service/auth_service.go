package service

import (
	"errors"

	"github.com/yousefi-abolfazl/code-judge/internal/auth"
	"github.com/yousefi-abolfazl/code-judge/internal/models"
	"github.com/yousefi-abolfazl/code-judge/internal/repository"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(username, password string) (*models.User, error) {
	existingUser, err := s.userRepo.FindByUsername(username)
	if err == nil && existingUser != nil {
		return nil, errors.New("username already exists")
	}
	
	user := &models.User{
		Username: username,
		Password: password,
		Role:     models.RoleUser,
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(username, password string) (string, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("invalid username or password")
		}
		return "", err
	}

	if !user.CheckPassword(password) {
		return "", errors.New("invalid username or password")
	}

	token, err := auth.GenerateToken(user, s.jwtSecret)
	if err != nil {
		return "", err
	}

	return token, nil
}
