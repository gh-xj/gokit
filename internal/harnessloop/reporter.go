package harnessloop

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
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
	if err := writeJSON(filepath.Join(dir, "findings.json"), result.Findings); err != nil {
		return err
	}
	ts := time.Now().UTC().Format("20060102-150405")
	reportPath := filepath.Join(dir, ts+"-report.md")
	body := fmt.Sprintf("# Onboarding Loop Report\n\n- Scenario: %s\n- Score: %.2f/10\n- Threshold: %.2f\n- Pass: %v\n- Iterations: %d\n- Branch: %s\n\n## Findings\n", result.Scenario.Name, result.Judge.Score, result.Judge.Threshold, result.Judge.Pass, result.Iterations, result.Branch)
	for _, f := range result.Findings {
		body += fmt.Sprintf("- [%s] %s (%s)\n", f.Code, f.Message, f.Source)
	}
	if err := os.WriteFile(reportPath, []byte(body), 0644); err != nil {
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
