package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/models"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/repository"
)

type RunnerHandler struct {
	submissionRepo *repository.SubmissionRepository
	problemRepo    *repository.ProblemRepository
}

func NewRunnerHandler(
	submissionRepo *repository.SubmissionRepository,
	problemRepo *repository.ProblemRepository,
) *RunnerHandler {
	return &RunnerHandler{
		submissionRepo: submissionRepo,
		problemRepo:    problemRepo,
	}
}

// GetNextSubmission returns the next pending submission
// This is an internal API for the runner service
func (h *RunnerHandler) GetNextSubmission(c *gin.Context) {
	submission, err := h.submissionRepo.GetNextPendingSubmission()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get next submission"})
		return
	}

	if submission == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No pending submissions found"})
		return
	}

	// Get the problem details
	problem, err := h.problemRepo.GetProblemByID(submission.ProblemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get problem details"})
		return
	}

	if problem == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Problem not found"})
		return
	}

	// Return submission details with problem
	c.JSON(http.StatusOK, gin.H{
		"submission": submission,
		"problem":    problem,
	})
}

// UpdateSubmissionResult updates the result of a submission
// This is an internal API for the runner service
func (h *RunnerHandler) UpdateSubmissionResult(c *gin.Context) {
	submissionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid submission ID"})
		return
	}

	var result struct {
		Status        models.SubmissionStatus `json:"status"`
		Result        models.SubmissionResult `json:"result"`
		ExecutionTime int                     `json:"execution_time"`
		MemoryUsed    int                     `json:"memory_used"`
		ErrorMessage  string                  `json:"error_message"`
	}

	if err := c.BindJSON(&result); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get the submission
	submission, err := h.submissionRepo.GetSubmissionByID(uint(submissionID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get submission"})
		return
	}

	if submission == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	// Update submission with results
	submission.Status = result.Status
	submission.Result = result.Result
	submission.ExecutionTime = result.ExecutionTime
	submission.MemoryUsed = result.MemoryUsed
	submission.ErrorMessage = result.ErrorMessage

	if err := h.submissionRepo.UpdateSubmission(submission); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update submission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Submission updated successfully"})
}
