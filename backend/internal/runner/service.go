package runner

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/models"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/repository"
)

type RunnerService struct {
	submissionRepo  *repository.SubmissionRepository
	problemRepo     *repository.ProblemRepository
	runner          *Runner
	runnerTimeout   time.Duration
	maxRetries      int
	pendingLock     sync.Mutex
	processingTasks map[uint]struct{}
	logger          *logrus.Logger
}

func NewRunnerService(
	submissionRepo *repository.SubmissionRepository,
	problemRepo *repository.ProblemRepository,
	runner *Runner,
	runnerTimeout time.Duration,
	maxRetries int,
	logger *logrus.Logger,
) *RunnerService {
	if logger == nil {
		logger = logrus.New()
	}

	if runnerTimeout == 0 {
		runnerTimeout = 30 * time.Second
	}

	if maxRetries == 0 {
		maxRetries = 3
	}

	return &RunnerService{
		submissionRepo:  submissionRepo,
		problemRepo:     problemRepo,
		runner:          runner,
		runnerTimeout:   runnerTimeout,
		maxRetries:      maxRetries,
		processingTasks: make(map[uint]struct{}),
		logger:          logger,
	}
}

// GetNextPendingSubmission gets the next submission for processing
func (s *RunnerService) GetNextPendingSubmission() (*models.Submission, error) {
	s.pendingLock.Lock()
	defer s.pendingLock.Unlock()

	// Get the next pending submission
	submission, err := s.submissionRepo.GetNextPendingSubmission()
	if err != nil {
		return nil, fmt.Errorf("failed to get next pending submission: %w", err)
	}

	if submission == nil {
		return nil, nil // No pending submissions
	}

	// Check if the submission is already being processed
	if _, exists := s.processingTasks[submission.ID]; exists {
		return nil, nil // Another runner is processing this submission
	}

	// Mark the submission as processing
	submission.Status = models.StatusProcessing
	if err := s.submissionRepo.UpdateSubmission(submission); err != nil {
		return nil, fmt.Errorf("failed to update submission status: %w", err)
	}

	// Add to processing tasks
	s.processingTasks[submission.ID] = struct{}{}

	return submission, nil
}

// ProcessSubmission processes a submission
func (s *RunnerService) ProcessSubmission(submission *models.Submission) error {
	defer func() {
		// Remove from processing tasks when done
		s.pendingLock.Lock()
		delete(s.processingTasks, submission.ID)
		s.pendingLock.Unlock()
	}()

	// Get the problem
	problem, err := s.problemRepo.GetProblemByID(submission.ProblemID)
	if err != nil {
		submission.Status = models.StatusError
		submission.ErrorMessage = "Failed to retrieve problem details"
		s.submissionRepo.UpdateSubmission(submission)
		return fmt.Errorf("failed to get problem: %w", err)
	}

	// Run the code
	result, err := s.runner.RunCode(submission.Code, *problem)
	if err != nil {
		submission.Status = models.StatusError
		submission.ErrorMessage = "Runner error: " + err.Error()
		s.submissionRepo.UpdateSubmission(submission)
		return fmt.Errorf("runner error: %w", err)
	}

	// Update submission with results
	submission.Status = result.Status
	submission.Result = result.Result
	submission.ExecutionTime = result.ExecutionTime
	submission.MemoryUsed = result.MemoryUsed
	submission.ErrorMessage = result.ErrorMessage

	if err := s.submissionRepo.UpdateSubmission(submission); err != nil {
		return fmt.Errorf("failed to update submission with results: %w", err)
	}

	return nil
}

// StartProcessing starts processing submissions in a loop
func (s *RunnerService) StartProcessing(stopCh <-chan struct{}) {
	s.logger.Info("Starting submission processor")

	for {
		select {
		case <-stopCh:
			s.logger.Info("Stopping submission processor")
			return
		default:
			submission, err := s.GetNextPendingSubmission()
			if err != nil {
				s.logger.WithError(err).Error("Failed to get pending submission")
				time.Sleep(1 * time.Second)
				continue
			}

			if submission == nil {
				// No submissions to process, sleep for a bit
				time.Sleep(500 * time.Millisecond)
				continue
			}

			s.logger.WithField("submission_id", submission.ID).Info("Processing submission")

			// Process the submission in a separate goroutine with timeout
			processDone := make(chan error, 1)
			go func() {
				processDone <- s.ProcessSubmission(submission)
			}()

			// Wait for processing to complete or timeout
			select {
			case err := <-processDone:
				if err != nil {
					s.logger.WithError(err).WithField("submission_id", submission.ID).Error("Failed to process submission")
				} else {
					s.logger.WithField("submission_id", submission.ID).Info("Submission processed successfully")
				}
			case <-time.After(s.runnerTimeout):
				// Processing took too long, mark as error and continue
				s.logger.WithField("submission_id", submission.ID).Warn("Submission processing timed out")

				submission.Status = models.StatusError
				submission.ErrorMessage = "Submission processing timed out"
				s.submissionRepo.UpdateSubmission(submission)

				// Remove from processing tasks
				s.pendingLock.Lock()
				delete(s.processingTasks, submission.ID)
				s.pendingLock.Unlock()
			}
		}
	}
}
