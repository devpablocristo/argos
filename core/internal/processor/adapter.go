package processor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type Processor interface {
	ProcessCaptureGroups(ctx context.Context, inputPath, outputPath string) (Response, error)
}

type CLIConfig struct {
	Python     string
	PythonPath string
	Timeout    time.Duration
}

type CLIProcessor struct {
	cfg CLIConfig
}

func NewCLIProcessor(cfg CLIConfig) *CLIProcessor {
	if cfg.Python == "" {
		cfg.Python = "python3"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 2 * time.Minute
	}
	return &CLIProcessor{cfg: cfg}
}

func (p *CLIProcessor) ProcessCaptureGroups(ctx context.Context, inputPath, outputPath string) (Response, error) {
	ctx, cancel := context.WithTimeout(ctx, p.cfg.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, p.cfg.Python, "-m", "argos_processing", "process-capture-group", "--input", inputPath, "--output", outputPath)
	cmd.Env = os.Environ()
	if p.cfg.PythonPath != "" {
		cmd.Env = append(cmd.Env, "PYTHONPATH="+p.cfg.PythonPath)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return Response{}, fmt.Errorf("run processing worker: %w: %s", err, stderr.String())
	}
	var response Response
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return Response{}, fmt.Errorf("decode processing response: %w", err)
	}
	return response, nil
}
