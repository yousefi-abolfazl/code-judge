package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/models"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/service"
)

type ProblemHandler struct {
	problemService *service.ProblemService
}

func NewProblemHandler(problemService *service.ProblemService) *ProblemHandler {
	return &ProblemHandler{
		problemService: problemService,
	}
}

type CreateProblemRequest struct {
	Title       string `json:"title" binding:"required"`
	Statement   string `json:"statement" binding:"required"`
	TimeLimit   int    `json:"time_limit" binding:"required"`
	MemoryLimit int    `json:"memory_limit" binding:"required"`
	Input       string `json:"input" binding:"required"`
	Output      string `json:"output" binding:"required"`
}

func (h *ProblemHandler) CreateProblem(c *gin.Context) {
	var req CreateProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	problem, err := h.problemService.CreateProblem(
		req.Title, req.Statement, req.Input, req.Output,
		req.TimeLimit, req.MemoryLimit, userID.(uint),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create problem"})
		return
	}

	c.JSON(http.StatusCreated, problem)
}

func (h *ProblemHandler) GetPublishedProblems(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	problems, total, err := h.problemService.GetPublishedProblems(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve problems"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"problems": problems,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func (h *ProblemHandler) GetProblemByID(c *gin.Context) {
	problemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid problem ID"})
		return
	}

	problem, err := h.problemService.GetProblemByID(uint(problemID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve problem"})
		return
	}

	if problem == nil || problem.Status != models.ProblemPublished {
		c.JSON(http.StatusNotFound, gin.H{"error": "Problem not found or not published"})
		return
	}

	c.JSON(http.StatusOK, problem)
}

func (h *ProblemHandler) UpdateProblemStatus(c *gin.Context) {
	problemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid problem ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var serviceErr error
	if req.Status == string(models.ProblemPublished) {
		serviceErr = h.problemService.PublishProblem(uint(problemID))
	} else if req.Status == string(models.ProblemDraft) {
		serviceErr = h.problemService.UnpublishProblem(uint(problemID))
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status value"})
		return
	}

	if serviceErr != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": serviceErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Problem status updated successfully"})
}

// GetAllProblems handles listing all problems for admins
func (h *ProblemHandler) GetAllProblems(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	problems, total, err := h.problemService.GetAllProblems(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve all problems"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"problems": problems,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}
