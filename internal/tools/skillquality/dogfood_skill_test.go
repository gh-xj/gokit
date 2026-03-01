package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDogfoodSkillFrontmatterHasRequiredKeys(t *testing.T) {
	_, content := loadDogfoodSkill(t)
	frontmatter, ok := parseDogfoodFrontmatter(content)
	if !ok {
		t.Fatalf("expected YAML frontmatter in dogfood skill doc")
	}

	for _, key := range []string{"name", "description", "version"} {
		value := strings.TrimSpace(frontmatter[key])
		if value == "" {
			t.Fatalf("frontmatter missing required key %q", key)
		}
	}

	if got := strings.TrimSpace(frontmatter["name"]); got != "dogfood-feedback-loop" {
		t.Fatalf("unexpected frontmatter name: got %q want %q", got, "dogfood-feedback-loop")
	}
}

func TestDogfoodSkillHasRequiredSections(t *testing.T) {
	_, content := loadDogfoodSkill(t)

	for _, heading := range []string{
		"## In scope",
		"## Use this when",
		"## Feedback workflow",
		"## Replay",
	} {
		if !strings.Contains(content, heading) {
			t.Fatalf("missing required section heading %q", heading)
		}
	}
}

func TestDogfoodSkillExamplesExist(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	for _, relPath := range []string{
		filepath.Join("skills", "dogfood-feedback-loop", "examples", "local-runtime-error.md"),
		filepath.Join("skills", "dogfood-feedback-loop", "examples", "ci-failure.md"),
	} {
		absPath := filepath.Join(root, relPath)
		info, err := os.Stat(absPath)
		if err != nil {
			t.Fatalf("expected example file %q to exist: %v", absPath, err)
		}
		if info.IsDir() {
			t.Fatalf("expected example file %q, got directory", absPath)
		}
	}
}

func loadDogfoodSkill(t *testing.T) (string, string) {
	t.Helper()
	root := filepath.Join("..", "..", "..")
	skillPath := filepath.Join(root, "skills", "dogfood-feedback-loop", "SKILL.md")
	raw, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read skill doc: %v", err)
	}
	return skillPath, string(raw)
}

func parseDogfoodFrontmatter(content string) (map[string]string, bool) {
	lines := strings.Split(content, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return nil, false
	}

	fm := make(map[string]string)
	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			return fm, true
		}
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
		fm[key] = value
	}
	return nil, false
}
