package repository

import (
	"github.com/yousefi-abolfazl/code-judge/backend/internal/models"
	"gorm.io/gorm"
)

type SubmissionRepository struct {
	db *gorm.DB
}

func NewSubmissionRepository(db *gorm.DB) *SubmissionRepository {
	return &SubmissionRepository{db}
}

// CreateSubmission creates a new submission
func (r *SubmissionRepository) CreateSubmission(submission *models.Submission) error {
	return r.db.Create(submission).Error
}

// GetSubmissionByID gets a submission by ID
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

// GetSubmissionsByUserID gets all submissions for a user
func (r *SubmissionRepository) GetSubmissionsByUserID(userID uint, page, pageSize int) ([]models.Submission, int64, error) {
	var submissions []models.Submission
	var count int64

	// Count total submissions for pagination
	if err := r.db.Model(&models.Submission{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Get submissions with pagination
	offset := (page - 1) * pageSize
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&submissions).Error; err != nil {
		return nil, 0, err
	}

	return submissions, count, nil
}

// GetSubmissionsByProblemID gets all submissions for a problem
func (r *SubmissionRepository) GetSubmissionsByProblemID(problemID uint, page, pageSize int) ([]models.Submission, int64, error) {
	var submissions []models.Submission
	var count int64

	// Count total submissions for pagination
	if err := r.db.Model(&models.Submission{}).Where("problem_id = ?", problemID).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Get submissions with pagination
	offset := (page - 1) * pageSize
	if err := r.db.Where("problem_id = ?", problemID).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&submissions).Error; err != nil {
		return nil, 0, err
	}

	return submissions, count, nil
}

// UpdateSubmission updates a submission
func (r *SubmissionRepository) UpdateSubmission(submission *models.Submission) error {
	return r.db.Save(submission).Error
}

// GetNextPendingSubmission gets the next pending submission using FOR UPDATE to prevent race conditions
func (r *SubmissionRepository) GetNextPendingSubmission() (*models.Submission, error) {
	var submission models.Submission

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Get the oldest pending submission with a lock to prevent other processes from selecting the same row
		result := tx.Exec("SELECT * FROM submissions WHERE status = ? ORDER BY created_at ASC LIMIT 1 FOR UPDATE", models.StatusPending).Scan(&submission)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return gorm.ErrRecordNotFound
			}
			return result.Error
		}

		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &submission, nil
}
