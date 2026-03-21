package strategy

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//go:embed defaults/*
var defaultsFS embed.FS

// Discover walks up from startDir looking for .agentops/ and loads the strategy.
func Discover(startDir string) (*Strategy, error) {
	root, err := findRoot(startDir)
	if err != nil {
		return nil, err
	}
	return load(root)
}

// Bootstrap creates .agentops/ with default files. Idempotent: does not overwrite existing files.
func Bootstrap(projectDir string) error {
	agentopsDir := filepath.Join(projectDir, ".agentops")
	if err := os.MkdirAll(agentopsDir, 0o755); err != nil {
		return fmt.Errorf("create .agentops/: %w", err)
	}

	entries, err := defaultsFS.ReadDir("defaults")
	if err != nil {
		return fmt.Errorf("read embedded defaults: %w", err)
	}

	for _, entry := range entries {
		target := filepath.Join(agentopsDir, entry.Name())
		if _, err := os.Stat(target); err == nil {
			continue // don't overwrite existing
		}
		data, err := defaultsFS.ReadFile("defaults/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read default %s: %w", entry.Name(), err)
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", entry.Name(), err)
		}
	}
	return nil
}

func findRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	for {
		if info, err := os.Stat(filepath.Join(dir, ".agentops")); err == nil && info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .agentops/ found (searched up from %s)", startDir)
		}
		dir = parent
	}
}

func load(root string) (*Strategy, error) {
	agentopsDir := filepath.Join(root, ".agentops")
	s := &Strategy{Root: root}

	// Load storage.yaml
	if err := loadYAML(filepath.Join(agentopsDir, "storage.yaml"), &s.Storage); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("storage.yaml: %w", err)
	}

	// Load transitions.yaml
	if err := loadYAML(filepath.Join(agentopsDir, "transitions.yaml"), &s.Transitions); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("transitions.yaml: %w", err)
	}

	// Load risk.yaml (unstructured)
	if err := loadYAML(filepath.Join(agentopsDir, "risk.yaml"), &s.Risk); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("risk.yaml: %w", err)
	}

	// Load routing.yaml (unstructured)
	if err := loadYAML(filepath.Join(agentopsDir, "routing.yaml"), &s.Routing); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("routing.yaml: %w", err)
	}

	// Load budget.yaml (unstructured)
	if err := loadYAML(filepath.Join(agentopsDir, "budget.yaml"), &s.Budget); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("budget.yaml: %w", err)
	}

	// Load hooks.yaml
	if err := loadYAML(filepath.Join(agentopsDir, "hooks.yaml"), &s.Hooks); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("hooks.yaml: %w", err)
	}

	// Load schema.md (raw)
	if data, err := os.ReadFile(filepath.Join(agentopsDir, "schema.md")); err == nil {
		s.SchemaTemplate = string(data)
	}

	return s, nil
}

func loadYAML(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, target)
}
