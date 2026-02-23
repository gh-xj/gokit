package harnessloop

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func WriteReports(repoRoot string, result RunResult) error {
	dir := filepath.Join(repoRoot, ".docs", "onboarding-loop")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	latest := filepath.Join(dir, "latest-summary.json")
	if err := writeJSON(latest, result); err != nil {
		return err
	}
	return nil
}

func writeJSON(path string, v any) error {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0644)
}
