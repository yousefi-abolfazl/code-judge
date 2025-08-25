package service

import (
	"github.com/yousefi-abolfazl/code-judge/backend/internal/models"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/repository"
)

type ProblemService struct {
	problemRepo *repository.ProblemRepository
}

func NewProblemService(problemRepo *repository.ProblemRepository) *ProblemService {
	return &ProblemService{
		problemRepo: problemRepo,
	}
}

func (s *ProblemService) CreateProblem(title, statement, input, output string, timeLimit, memoryLimit int, ownerID uint) (*models.Problem, error) {
	problem := &models.Problem{
		Title:       title,
		Statement:   statement,
		TimeLimit:   timeLimit,
		MemoryLimit: memoryLimit,
		Input:       input,
		Output:      output,
		OwnerID:     ownerID,
		Status:      models.ProblemDraft,
	}

	err := s.problemRepo.CreateProblem(problem)
	if err != nil {
		return nil, err
	}
	return problem, nil
}

func (s *ProblemService) GetProblemByID(id uint) (*models.Problem, error) {
	return s.problemRepo.GetProblemByID(id)
}

func (s *ProblemService) GetPublishedProblems(page, pageSize int) ([]models.Problem, int64, error) {
	return s.problemRepo.GetPublishedProblems(page, pageSize)
}
