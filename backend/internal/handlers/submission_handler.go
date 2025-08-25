package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/service"
)

type SubmissionHandler struct {
	submissionService *service.SubmissionService
}

func NewSubmissionHandler(submissionService *service.SubmissionService) *SubmissionHandler {
	return &SubmissionHandler{
		submissionService: submissionService,
	}
}

type CreateSubmissionRequest struct {
	Code      string `json:"code" binding:"required"`
	ProblemID uint   `json:"problem_id" binding:"required"`
}

func (h *SubmissionHandler) CreateSubmission(c *gin.Context) {
	var req CreateSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	submission, err := h.submissionService.CreateSubmission(req.Code, req.ProblemID, userID.(uint))
	if err != nil {
		if err.Error() == "problem not found or is not published" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create submission"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Submission received successfully. It will be processed shortly.",
		"submission": submission,
	})
}

func (h *SubmissionHandler) GetMySubmissions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	submissions, total, err := h.submissionService.GetSubmissionsByUserID(userID.(uint), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve submissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"submissions": submissions,
		"total":       total,
		"page":        page,
		"pageSize":    pageSize,
	})
}
