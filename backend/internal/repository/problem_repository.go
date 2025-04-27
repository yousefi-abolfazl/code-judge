package repository

import (
	"time"

	"github.com/yousefi-abolfazl/code-judge/backend/internal/models"
	"gorm.io/gorm"
)

type ProblemRepository struct {
	db *gorm.DB
}

func NewProblemRepository(db *gorm.DB) *ProblemRepository {
	return &ProblemRepository{db}
}

// CreateProblem creates a new problem
func (r *ProblemRepository) CreateProblem(problem *models.Problem) error {
	return r.db.Create(problem).Error
}

// GetProblemByID gets a problem by ID
func (r *ProblemRepository) GetProblemByID(id uint) (*models.Problem, error) {
	var problem models.Problem
	if err := r.db.First(&problem, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &problem, nil
}

// GetPublishedProblems gets all published problems with pagination
func (r *ProblemRepository) GetPublishedProblems(page, pageSize int) ([]models.Problem, int64, error) {
	var problems []models.Problem
	var count int64

	// Count total published problems
	if err := r.db.Model(&models.Problem{}).Where("status = ?", models.ProblemPublished).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Get problems with pagination
	offset := (page - 1) * pageSize
	if err := r.db.Where("status = ?", models.ProblemPublished).
		Order("publish_date DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&problems).Error; err != nil {
		return nil, 0, err
	}

	return problems, count, nil
}

// GetProblemsByOwnerID gets all problems created by a user
func (r *ProblemRepository) GetProblemsByOwnerID(ownerID uint) ([]models.Problem, error) {
	var problems []models.Problem
	if err := r.db.Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&problems).Error; err != nil {
		return nil, err
	}
	return problems, nil
}

// GetAllProblems gets all problems (for admin)
func (r *ProblemRepository) GetAllProblems(page, pageSize int) ([]models.Problem, int64, error) {
	var problems []models.Problem
	var count int64

	// Count total problems
	if err := r.db.Model(&models.Problem{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Get problems with pagination
	offset := (page - 1) * pageSize
	if err := r.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&problems).Error; err != nil {
		return nil, 0, err
	}

	return problems, count, nil
}

// UpdateProblem updates a problem
func (r *ProblemRepository) UpdateProblem(problem *models.Problem) error {
	return r.db.Save(problem).Error
}

// PublishProblem changes a problem's status to published
func (r *ProblemRepository) PublishProblem(id uint) error {
	now := time.Now()
	return r.db.Model(&models.Problem{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       models.ProblemPublished,
		"publish_date": now,
	}).Error
}

// UnpublishProblem changes a problem's status to draft
func (r *ProblemRepository) UnpublishProblem(id uint) error {
	return r.db.Model(&models.Problem{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       models.ProblemDraft,
		"publish_date": nil,
	}).Error
}
