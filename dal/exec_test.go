package dal

import (
	"strings"
	"testing"
)

func TestExecutorImpl_Run(t *testing.T) {
	ex := NewExecutor()
	out, err := ex.Run("echo", "hello")
	if err != nil {
		t.Fatalf("Run(echo hello) error: %v", err)
	}
	if got := strings.TrimSpace(out); got != "hello" {
		t.Errorf("Run(echo hello) = %q, want %q", got, "hello")
	}
}

func TestExecutorImpl_RunInDir(t *testing.T) {
	ex := NewExecutor()
	out, err := ex.RunInDir("/tmp", "pwd")
	if err != nil {
		t.Fatalf("RunInDir(/tmp, pwd) error: %v", err)
	}
	// /tmp may resolve to /private/tmp on macOS
	got := strings.TrimSpace(out)
	if got != "/tmp" && got != "/private/tmp" {
		t.Errorf("RunInDir(/tmp, pwd) = %q, want /tmp or /private/tmp", got)
	}
}

func TestExecutorImpl_Which(t *testing.T) {
	ex := NewExecutor()
	if !ex.Which("echo") {
		t.Error("Which(echo) = false, want true")
	}
	if ex.Which("nonexistent-command-xyz-999") {
		t.Error("Which(nonexistent) = true, want false")
	}
}
