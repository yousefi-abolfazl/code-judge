package runner

import (
	"time"

	"github.com/sirupsen/logrus"
)

// Processor manages the processing of submissions
type Processor struct {
	client     *APIClient
	runner     *Runner
	logger     *logrus.Logger
	retryDelay time.Duration
	maxRetries int
}

// NewProcessor creates a new processor
func NewProcessor(client *APIClient, runner *Runner, logger *logrus.Logger) *Processor {
	if logger == nil {
		logger = logrus.New()
	}

	return &Processor{
		client:     client,
		runner:     runner,
		logger:     logger,
		retryDelay: 500 * time.Millisecond,
		maxRetries: 3,
	}
}

// Start begins processing submissions in a loop
func (p *Processor) Start(stopCh <-chan struct{}) {
	p.logger.Info("Starting submission processor")

	for {
		select {
		case <-stopCh:
			p.logger.Info("Stopping submission processor")
			return
		default:
			p.processNextSubmission()

			// Brief pause between processing attempts to avoid hammering the API
			time.Sleep(p.retryDelay)
		}
	}
}

// processNextSubmission processes the next available submission
func (p *Processor) processNextSubmission() {
	// Get the next submission from the API
	task, err := p.client.GetNextSubmission()
	if err != nil {
		p.logger.WithError(err).Error("Failed to get next submission")
		return
	}

	if task == nil {
		// No pending submissions, nothing to do
		return
	}

	p.logger.WithField("submission_id", task.Submission.ID).Info("Processing submission")

	// Run the code
	result, err := p.runner.RunCode(task.Submission.Code, *task.Problem)
	if err != nil {
		p.logger.WithError(err).WithField("submission_id", task.Submission.ID).Error("Failed to run code")

		// Send error result back to API
		result = &RunResult{
			Status:       "error",
			Result:       "Runtime Error",
			ErrorMessage: "Failed to run code: " + err.Error(),
		}
	}

	// Send the result back to the API
	if err := p.client.SendSubmissionResult(task.Submission.ID, result); err != nil {
		p.logger.WithError(err).WithField("submission_id", task.Submission.ID).Error("Failed to send result to API")
	} else {
		p.logger.WithField("submission_id", task.Submission.ID).Info("Submission processed successfully")
	}
}
