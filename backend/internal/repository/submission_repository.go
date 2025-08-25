package repository

import (
	"errors"

	"github.com/yousefi-abolfazl/code-judge/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SubmissionRepository struct {
	db *gorm.DB
}

func NewSubmissionRepository(db *gorm.DB) *SubmissionRepository {
	return &SubmissionRepository{db}
}

func (r *SubmissionRepository) CreateSubmission(submission *models.Submission) error {
	return r.db.Create(submission).Error
}

func (r *SubmissionRepository) GetSubmissionByID(id uint) (*models.Submission, error) {
	var submission models.Submission
	if err := r.db.First(&submission, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &submission, nil
}

func (r *SubmissionRepository) GetSubmissionsByUserID(userID uint, page, pageSize int) ([]models.Submission, int64, error) {
	var submissions []models.Submission
	var count int64

	if err := r.db.Model(&models.Submission{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&submissions).Error; err != nil {
		return nil, 0, err
	}

	return submissions, count, nil
}

func (r *SubmissionRepository) GetSubmissionsByProblemID(problemID uint, page, pageSize int) ([]models.Submission, int64, error) {
	var submissions []models.Submission
	var count int64

	if err := r.db.Model(&models.Submission{}).Where("problem_id = ?", problemID).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := r.db.Where("problem_id = ?", problemID).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&submissions).Error; err != nil {
		return nil, 0, err
	}

	return submissions, count, nil
}

func (r *SubmissionRepository) UpdateSubmission(submission *models.Submission) error {
	return r.db.Save(submission).Error
}

func (r *SubmissionRepository) GetNextPendingSubmission() (*models.Submission, error) {
	var submission models.Submission

	err := r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("status = ?", models.StatusPending).
			Order("created_at ASC").
			First(&submission).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return gorm.ErrRecordNotFound
			}
			return err
		}

		submission.Status = models.StatusProcessing
		if err := tx.Save(&submission).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &submission, nil
}
