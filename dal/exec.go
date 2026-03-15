package dal

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// ExecutorImpl is the real OS-backed Executor.
type ExecutorImpl struct{}

// NewExecutor returns a new ExecutorImpl.
func NewExecutor() *ExecutorImpl { return &ExecutorImpl{} }

func (e *ExecutorImpl) Run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

func (e *ExecutorImpl) RunInDir(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

func (e *ExecutorImpl) RunOsascript(script string) string {
	cmd := exec.Command("osascript", "-e", script)
	out, _ := cmd.Output()
	return strings.TrimSpace(string(out))
}

func (e *ExecutorImpl) Which(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
