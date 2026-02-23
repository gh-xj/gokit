package harnessloop

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReplayIteration(t *testing.T) {
	repo := t.TempDir()
	iterDir := filepath.Join(repo, ".docs", "onboarding-loop", "runs", "r1", "iter-01")
	if err := os.MkdirAll(iterDir, 0755); err != nil {
		t.Fatalf("mkdir iter dir: %v", err)
	}
	ctx := map[string]any{
		"scenario": map[string]any{
			"name":  "seed",
			"steps": []map[string]any{{"name": "ok", "command": "echo ok"}},
		},
	}
	if err := writeJSON(filepath.Join(iterDir, "planner-context.json"), ctx); err != nil {
		t.Fatalf("write context: %v", err)
	}
	report, err := ReplayIteration(repo, "r1", 1, 9.0)
	if err != nil {
		t.Fatalf("replay iteration: %v", err)
	}
	if !report.Replay.OK {
		t.Fatalf("expected replay ok")
	}
}
