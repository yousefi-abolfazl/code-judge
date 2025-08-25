package repository

import (
	"github.com/yousefi-abolfazl/code-judge/backend/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

type UserStats struct {
	TotalSubmissions      int64 `json:"total_submissions"`
	SuccessfulSubmissions int64 `json:"successful_submissions"`
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUserStats(userID uint) (*UserStats, error) {
	var stats UserStats

	// Get total submissions
	err := r.db.Model(&models.Submission{}).Where("user_id = ?", userID).Count(&stats.TotalSubmissions).Error
	if err != nil {
		return nil, err
	}

	// Get successful submissions
	err = r.db.Model(&models.Submission{}).
		Where("user_id = ? AND result = ?", userID, models.ResultOK).
		Count(&stats.SuccessfulSubmissions).Error
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (r *UserRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateRole(userID uint, role string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("role", role).Error
}
