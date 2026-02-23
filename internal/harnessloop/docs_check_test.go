package harnessloop

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckOnboardingInstallReadinessDetectsMissingInstall(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("AI Prompt Starter"), 0644); err != nil {
		t.Fatalf("write readme: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "prompts"), 0755); err != nil {
		t.Fatalf("mkdir prompts: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "prompts", "agentcli-onboarding.prompt.md"), []byte("no install here"), 0644); err != nil {
		t.Fatalf("write prompt: %v", err)
	}
	findings := CheckOnboardingInstallReadiness(root)
	if len(findings) == 0 {
		t.Fatal("expected findings for missing install readiness")
	}
}
