package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDogfoodSkillHasRequiredSections(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	skillPath := filepath.Join(root, "skills", "dogfood-feedback-loop", "SKILL.md")

	raw, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read skill doc: %v", err)
	}
	content := string(raw)

	requiredSnippets := []string{
		"name: dogfood-feedback-loop",
		"description:",
		"version:",
		"## In scope",
		"## Use this when",
		"## Feedback workflow",
		"## Replay",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(content, snippet) {
			t.Fatalf("missing required snippet %q in %s", snippet, skillPath)
		}
	}
}
