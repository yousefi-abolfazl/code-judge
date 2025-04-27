package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/models"
)

type Runner struct {
	dockerClient *client.Client
	tempDir      string
}

type RunResult struct {
	Status        models.SubmissionStatus
	Result        models.SubmissionResult
	ExecutionTime int
	MemoryUsed    int
	ErrorMessage  string
}

// NewRunner creates a new code runner instance
func NewRunner(tempDir string) (*Runner, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	if tempDir == "" {
		tempDir = os.TempDir()
	}

	// Create temp directory if it doesn't exist
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create temp directory: %w", err)
		}
	}

	return &Runner{
		dockerClient: cli,
		tempDir:      tempDir,
	}, nil
}

// RunCode runs the submitted code in a Docker container with strict limitations
func (r *Runner) RunCode(code string, problem models.Problem) (*RunResult, error) {
	result := &RunResult{
		Status: models.StatusProcessing,
	}

	// Create a unique directory for this submission
	submissionID := uuid.New().String()
	submissionDir := filepath.Join(r.tempDir, submissionID)
	if err := os.MkdirAll(submissionDir, 0755); err != nil {
		result.Status = models.StatusError
		result.ErrorMessage = "Failed to create submission directory"
		return result, fmt.Errorf("failed to create submission directory: %w", err)
	}
	defer os.RemoveAll(submissionDir) // Clean up when done

	// Write code to a file
	codeFile := filepath.Join(submissionDir, "main.go")
	if err := os.WriteFile(codeFile, []byte(code), 0644); err != nil {
		result.Status = models.StatusError
		result.ErrorMessage = "Failed to write code file"
		return result, fmt.Errorf("failed to write code file: %w", err)
	}

	// Write input to a file
	inputFile := filepath.Join(submissionDir, "input.txt")
	if err := os.WriteFile(inputFile, []byte(problem.Input), 0644); err != nil {
		result.Status = models.StatusError
		result.ErrorMessage = "Failed to write input file"
		return result, fmt.Errorf("failed to write input file: %w", err)
	}

	// Create a configuration file for the runner script
	configFile := filepath.Join(submissionDir, "config.json")
	configContent := fmt.Sprintf(`{
		"time_limit": %d,
		"memory_limit": %d
	}`, problem.TimeLimit, problem.MemoryLimit)
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		result.Status = models.StatusError
		result.ErrorMessage = "Failed to write config file"
		return result, fmt.Errorf("failed to write config file: %w", err)
	}

	// Create output file
	outputFile := filepath.Join(submissionDir, "output.txt")
	if _, err := os.Create(outputFile); err != nil {
		result.Status = models.StatusError
		result.ErrorMessage = "Failed to create output file"
		return result, fmt.Errorf("failed to create output file: %w", err)
	}

	// Build the Docker image if needed (this could be done once at startup)
	// For simplicity, assuming the image 'code-judge-runner' already exists

	// Run the code in a container
	ctx := context.Background()
	resp, err := r.dockerClient.ContainerCreate(
		ctx,
		&container.Config{
			Image: "golang:1.24-alpine",
			Cmd: []string{
				"sh", "-c",
				"cd /app && go run main.go < input.txt > output.txt 2> error.txt",
			},
			WorkingDir: "/app",
			Tty:        false,
		},
		&container.HostConfig{
			Binds: []string{
				fmt.Sprintf("%s:/app", submissionDir),
			},
			Resources: container.Resources{
				Memory:         int64(problem.MemoryLimit) * 1024 * 1024, // Convert MB to bytes
				CPUPeriod:      100000,
				CPUQuota:       100000,            // Use 1 CPU
				PidsLimit:      &[]int64{100}[0],  // Limit number of processes
				OomKillDisable: &[]bool{false}[0], // Allow OOM killer
			},
			NetworkMode: "none", // Disable network
		},
		nil, nil, "",
	)
	if err != nil {
		result.Status = models.StatusError
		result.ErrorMessage = "Failed to create container"
		return result, fmt.Errorf("failed to create container: %w", err)
	}
	defer r.dockerClient.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})

	// Start the container
	startTime := time.Now()
	if err := r.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		result.Status = models.StatusError
		result.ErrorMessage = "Failed to start container"
		return result, fmt.Errorf("failed to start container: %w", err)
	}

	// Set timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(problem.TimeLimit+1000)*time.Millisecond)
	defer cancel()

	// Wait for the container to finish
	statusCh, errCh := r.dockerClient.ContainerWait(timeoutCtx, resp.ID, container.WaitConditionNotRunning)
	var statusCode int64
	select {
	case err := <-errCh:
		if err != nil {
			result.Status = models.StatusError
			if timeoutCtx.Err() == context.DeadlineExceeded {
				// Container took too long to finish
				result.Result = models.ResultTimeLimit
				result.ErrorMessage = "Time limit exceeded"
			} else {
				result.ErrorMessage = fmt.Sprintf("Error waiting for container: %v", err)
			}
			return result, fmt.Errorf("error waiting for container: %w", err)
		}
	case status := <-statusCh:
		statusCode = status.StatusCode
	}

	executionTime := int(time.Since(startTime).Milliseconds())

	// Check if there were compilation/runtime errors
	errorContent, err := os.ReadFile(filepath.Join(submissionDir, "error.txt"))
	if err == nil && len(errorContent) > 0 {
		result.Status = models.StatusRejected

		if statusCode != 0 {
			// Runtime error
			result.Result = models.ResultRuntimeError
			result.ErrorMessage = string(errorContent)
		} else {
			// Check if there were compilation errors
			result.Result = models.ResultCompileError
			result.ErrorMessage = string(errorContent)
		}
		return result, nil
	}

	// Read the output
	outputContent, err := os.ReadFile(outputFile)
	if err != nil {
		result.Status = models.StatusError
		result.ErrorMessage = "Failed to read output file"
		return result, fmt.Errorf("failed to read output file: %w", err)
	}

	// Compare with expected output
	if string(outputContent) == problem.Output {
		result.Status = models.StatusAccepted
		result.Result = models.ResultOK
	} else {
		result.Status = models.StatusRejected
		result.Result = models.ResultWrongAnswer
		result.ErrorMessage = "Output does not match expected output"
	}

	// Get container stats for memory usage
	stats, err := r.dockerClient.ContainerStats(ctx, resp.ID, false)
	if err == nil {
		defer stats.Body.Close()
		var statsData map[string]interface{}
		if statsBody, err := io.ReadAll(stats.Body); err == nil {
			if err := json.Unmarshal(statsBody, &statsData); err == nil {
				if memStats, ok := statsData["memory_stats"].(map[string]interface{}); ok {
					if usage, ok := memStats["max_usage"].(float64); ok {
						// Convert bytes to MB
						result.MemoryUsed = int(usage / (1024 * 1024))
					}
				}
			}
		}
	}

	result.ExecutionTime = executionTime

	return result, nil
}
