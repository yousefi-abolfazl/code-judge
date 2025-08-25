// in internal/runner/runner.go

package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/yousefi-abolfazl/code-judge/backend/internal/models"
)

type Runner struct {
	dockerClient *client.Client
	tempDir      string
	dockerImage  string
}

// RunResult ساختار خروجی را کمی تغییر می‌دهیم تا با مدل‌ها هماهنگ‌تر باشد
type RunResult struct {
	Status        models.SubmissionStatus `json:"status"`
	Result        models.SubmissionResult `json:"result"`
	ExecutionTime int                     `json:"execution_time"` // in ms
	MemoryUsed    int                     `json:"memory_used"`    // in MB
	ErrorMessage  string                  `json:"error_message"`
}

// NewRunner حالا ایمیج داکر را هم به عنوان ورودی می‌گیرد
func NewRunner(tempDir string, dockerImage string) (*Runner, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	if tempDir == "" {
		tempDir = os.TempDir()
	}
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create temp directory: %w", err)
		}
	}

	return &Runner{
		dockerClient: cli,
		tempDir:      tempDir,
		dockerImage:  dockerImage,
	}, nil
}

func (r *Runner) RunCode(code string, problem models.Problem) (*RunResult, error) {
	ctx := context.Background()
	submissionID := uuid.New().String()
	submissionDir := filepath.Join(r.tempDir, submissionID)

	if err := os.MkdirAll(submissionDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create submission directory: %w", err)
	}
	defer os.RemoveAll(submissionDir)

	if err := os.WriteFile(filepath.Join(submissionDir, "main.go"), []byte(code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write code file: %w", err)
	}
	if err := os.WriteFile(filepath.Join(submissionDir, "input.txt"), []byte(problem.Input), 0644); err != nil {
		return nil, fmt.Errorf("failed to write input file: %w", err)
	}

	// --- مرحله ۱: کامپایل کد ---
	compileCmd := []string{"go", "build", "-o", "main", "main.go"}
	compileResult, err := r.runContainer(ctx, "compiler", compileCmd, submissionDir, problem.TimeLimit, problem.MemoryLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to run compile container: %w", err)
	}

	if compileResult.ExitCode != 0 {
		return &RunResult{
			Status:       models.StatusRejected,
			Result:       models.ResultCompileError,
			ErrorMessage: compileResult.Stderr,
		}, nil
	}

	// --- مرحله ۲: اجرای کد کامپایل شده ---
	runCmd := []string{"/bin/sh", "-c", "./main < input.txt"}
	runResult, err := r.runContainer(ctx, "executor", runCmd, submissionDir, problem.TimeLimit, problem.MemoryLimit)
	if err != nil {
		// اگر اینجا خطا رخ دهد، معمولا به معنی TLE است
		if err == context.DeadlineExceeded {
			return &RunResult{
				Status:        models.StatusRejected,
				Result:        models.ResultTimeLimit,
				ExecutionTime: problem.TimeLimit,
				ErrorMessage:  "Time Limit Exceeded",
			}, nil
		}
		return nil, fmt.Errorf("failed to run execution container: %w", err)
	}

	// --- مرحله ۳: تحلیل نتیجه اجرا ---
	finalResult := &RunResult{
		ExecutionTime: int(runResult.Duration.Milliseconds()),
	}

	// بررسی نتایج بر اساس کد خروج
	switch runResult.ExitCode {
	case 0: // اجرای موفق
		if strings.TrimSpace(runResult.Stdout) == strings.TrimSpace(problem.Output) {
			finalResult.Status = models.StatusAccepted
			finalResult.Result = models.ResultOK
		} else {
			finalResult.Status = models.StatusRejected
			finalResult.Result = models.ResultWrongAnswer
			finalResult.ErrorMessage = "Output does not match the expected output."
		}
	case 137: // کد استاندارد برای OOM Kill
		finalResult.Status = models.StatusRejected
		finalResult.Result = models.ResultMemoryLimit
		finalResult.ErrorMessage = "Memory Limit Exceeded"
	default: // سایر خطاها به عنوان خطای زمان اجرا در نظر گرفته می‌شوند
		finalResult.Status = models.StatusRejected
		finalResult.Result = models.ResultRuntimeError
		finalResult.ErrorMessage = runResult.Stderr
	}

	return finalResult, nil
}

type containerRunResult struct {
	Stdout   string
	Stderr   string
	ExitCode int64
	Duration time.Duration
}

// تابع کمکی برای اجرای یک دستور در یک کانتینر
func (r *Runner) runContainer(ctx context.Context, name string, cmd []string, dir string, timeLimit int, memLimit int) (*containerRunResult, error) {

	startTime := time.Now()

	resp, err := r.dockerClient.ContainerCreate(ctx,
		&container.Config{
			Image:      r.dockerImage,
			Cmd:        cmd,
			WorkingDir: "/app",
		},
		&container.HostConfig{
			Binds: []string{fmt.Sprintf("%s:/app", dir)},
			Resources: container.Resources{
				Memory: int64(memLimit) * 1024 * 1024, // MB to Bytes
				// CPU محدودیت‌های دیگر را هم می‌توان اینجا اضافه کرد
			},
			NetworkMode: "none",
		}, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create container %s: %w", name, err)
	}
	defer r.dockerClient.ContainerRemove(context.Background(), resp.ID, container.RemoveOptions{Force: true})

	if err := r.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container %s: %w", name, err)
	}

	// ایجاد یک context با timeout برای این اجرا
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeLimit)*time.Millisecond)
	defer cancel()

	statusCh, errCh := r.dockerClient.ContainerWait(timeoutCtx, resp.ID, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		// اگر context.DeadlineExceeded رخ دهد، به معنی TLE است
		if err == context.DeadlineExceeded {
			return nil, context.DeadlineExceeded
		}
		return nil, err
	case status := <-statusCh:
		// خواندن لاگ‌ها (خروجی) از کانتینر
		out, err := r.dockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
		if err != nil {
			return nil, fmt.Errorf("failed to get container logs: %w", err)
		}
		defer out.Close()

		// خروجی داکر stdout و stderr را در یک استریم می‌دهد، ما باید آن‌ها را جدا کنیم
		// برای سادگی، فعلا همه را در یک بافر می‌خوانیم
		// در یک سیستم واقعی می‌توان از `stdcopy.StdCopy` برای جداسازی استفاده کرد.
		var stdout, stderr strings.Builder
		// This is a simplified log reading. For production, use `stdcopy.StdCopy`.
		logBytes, _ := io.ReadAll(out)
		// Heuristically split logs, assuming error logs contain "error" or "panic"
		if strings.Contains(string(logBytes), "error") || strings.Contains(string(logBytes), "panic") {
			stderr.Write(logBytes)
		} else {
			stdout.Write(logBytes)
		}

		return &containerRunResult{
			ExitCode: status.StatusCode,
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			Duration: time.Since(startTime),
		}, nil
	}
}
