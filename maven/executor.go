package maven

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"time"
)

// ExecutionResult represents the result of executing a Maven command
type ExecutionResult struct {
	Command   Command
	ExitCode  int
	Duration  time.Duration
	StartTime time.Time
	Output    []string
	Error     error
}

// OutputHandler is called for each line of output
type OutputHandler func(line string)

// Execute runs a Maven command and streams output
func Execute(ctx context.Context, cmd Command, workDir string, outputHandler OutputHandler) (*ExecutionResult, error) {
	result := &ExecutionResult{
		Command:   cmd,
		StartTime: time.Now(),
		Output:    []string{},
	}

	execCmd := exec.CommandContext(ctx, cmd.Executable, cmd.Args...)
	execCmd.Dir = workDir

	stdout, err := execCmd.StdoutPipe()
	if err != nil {
		result.Error = err
		return result, err
	}

	stderr, err := execCmd.StderrPipe()
	if err != nil {
		result.Error = err
		return result, err
	}

	if err := execCmd.Start(); err != nil {
		result.Error = err
		return result, err
	}

	// Stream output
	go streamOutput(stdout, outputHandler, &result.Output)
	go streamOutput(stderr, outputHandler, &result.Output)

	err = execCmd.Wait()
	result.Duration = time.Since(result.StartTime)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.Error = err
		}
	} else {
		result.ExitCode = 0
	}

	return result, nil
}

func streamOutput(r io.Reader, handler OutputHandler, output *[]string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		*output = append(*output, line)
		if handler != nil {
			handler(line)
		}
	}
}
