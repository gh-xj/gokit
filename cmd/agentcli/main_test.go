package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	agentcli "github.com/gh-xj/agentcli-go"
)

func TestRunAddCommandWithDescription(t *testing.T) {
	root := t.TempDir()
	projectPath, err := agentcli.ScaffoldNew(root, "samplecli", "example.com/samplecli")
	if err != nil {
		t.Fatalf("ScaffoldNew failed: %v", err)
	}

	exitCode := run([]string{
		"add",
		"command",
		"--dir", projectPath,
		"--description", "sync files from source to target",
		"sync-data",
	})
	if exitCode != agentcli.ExitSuccess {
		t.Fatalf("unexpected exit code: got %d want %d", exitCode, agentcli.ExitSuccess)
	}

	content, err := os.ReadFile(filepath.Join(projectPath, "cmd", "sync-data.go"))
	if err != nil {
		t.Fatalf("read generated command file: %v", err)
	}
	if !strings.Contains(string(content), `Description: "sync files from source to target"`) {
		t.Fatalf("expected description in generated command file: %s", string(content))
	}
}

func TestRunAddCommandDescriptionRequiresValue(t *testing.T) {
	exitCode := run([]string{"add", "command", "--description"})
	if exitCode != agentcli.ExitUsage {
		t.Fatalf("unexpected exit code: got %d want %d", exitCode, agentcli.ExitUsage)
	}
}
