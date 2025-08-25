package service

import (
	"errors"

	"github.com/yousefi-abolfazl/code-judge/backend/internal/models"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/repository"
)

type SubmissionService struct {
	submissionRepo *repository.SubmissionRepository
	problemRepo    *repository.ProblemRepository
}

func NewSubmissionService(submissionRepo *repository.SubmissionRepository, problemRepo *repository.ProblemRepository) *SubmissionService {
	return &SubmissionService{
		submissionRepo: submissionRepo,
		problemRepo:    problemRepo,
	}
}

func (s *SubmissionService) CreateSubmission(code string, problemID, userID uint) (*models.Submission, error) {
	problem, err := s.problemRepo.GetProblemByID(problemID)
	if err != nil {
		return nil, errors.New("failed to find the problem")
	}
	if problem == nil || problem.Status != models.ProblemPublished {
		return nil, errors.New("problem not found or is not published")
	}

	submission := &models.Submission{
		UserID:    userID,
		ProblemID: problemID,
		Code:      code,
		Status:    models.StatusPending,
	}

	if err := s.submissionRepo.CreateSubmission(submission); err != nil {
		return nil, err
	}

	return submission, nil
}

func (s *SubmissionService) GetSubmissionsByUserID(userID uint, page, pageSize int) ([]models.Submission, int64, error) {
	return s.submissionRepo.GetSubmissionsByUserID(userID, page, pageSize)
}
