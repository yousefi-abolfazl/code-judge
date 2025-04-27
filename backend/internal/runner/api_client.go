package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/models"
)

// SubmissionTask represents a task received from the API
type SubmissionTask struct {
	Submission *models.Submission `json:"submission"`
	Problem    *models.Problem    `json:"problem"`
}

// APIClient handles communication with the backend API
type APIClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
	logger     *logrus.Logger
}

// NewAPIClient creates a new API client
func NewAPIClient(baseURL, token string, logger *logrus.Logger) *APIClient {
	if logger == nil {
		logger = logrus.New()
	}

	return &APIClient{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// GetNextSubmission retrieves the next pending submission from the API
func (c *APIClient) GetNextSubmission() (*SubmissionTask, error) {
	url := fmt.Sprintf("%s/submissions/next", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// No pending submissions
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s, status: %d", string(body), resp.StatusCode)
	}

	var response struct {
		Submission *models.Submission `json:"submission"`
		Problem    *models.Problem    `json:"problem"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &SubmissionTask{
		Submission: response.Submission,
		Problem:    response.Problem,
	}, nil
}

// SendSubmissionResult sends the result of a submission back to the API
func (c *APIClient) SendSubmissionResult(submissionID uint, result *RunResult) error {
	url := fmt.Sprintf("%s/submissions/%d/result", c.baseURL, submissionID)

	body, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s, status: %d", string(body), resp.StatusCode)
	}

	return nil
}
