package collect

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	collectorBinaryPath = "/opt/scrutiny/bin/scrutiny-collector-metrics"
	collectorTimeout    = 5 * time.Minute
)

type Status struct {
	Running      bool       `json:"running"`
	StartedAt    *time.Time `json:"startedAt"`
	FinishedAt   *time.Time `json:"finishedAt"`
	LastExitCode *int       `json:"lastExitCode"`
}

type Runner struct {
	mu     sync.Mutex
	status Status
	logger *logrus.Entry
}

func NewRunner(logger *logrus.Entry) *Runner {
	return &Runner{
		logger: logger.WithField("component", "collector-runner"),
	}
}

func (r *Runner) Start() (bool, error) {
	r.mu.Lock()
	if r.status.Running {
		r.mu.Unlock()
		return false, nil
	}

	start := time.Now().UTC()
	r.status.Running = true
	r.status.StartedAt = &start
	r.status.FinishedAt = nil
	r.mu.Unlock()

	go r.runCollector(start)

	return true, nil
}

func (r *Runner) GetStatus() Status {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.status
}

func (r *Runner) runCollector(startedAt time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), collectorTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, collectorBinaryPath, "run")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		exitCode = -1
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
	}

	if stdout.Len() > 0 {
		r.logger.Infof("collector stdout:\n%s", stdout.String())
	}
	if stderr.Len() > 0 {
		r.logger.Warnf("collector stderr:\n%s", stderr.String())
	}

	if ctx.Err() == context.DeadlineExceeded {
		r.logger.Errorf("collector timed out after %s", collectorTimeout)
	} else if err != nil {
		r.logger.Errorf("collector exited with error: %v", err)
	} else {
		r.logger.Info("collector finished successfully")
	}

	finishedAt := time.Now().UTC()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.status.Running = false
	r.status.FinishedAt = &finishedAt
	r.status.LastExitCode = &exitCode
	r.status.StartedAt = &startedAt
}
