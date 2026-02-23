package harnessloop

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type ReviewData struct {
	SchemaVersion string    `json:"schema_version"`
	RunID         string    `json:"run_id"`
	Mode          string    `json:"mode"`
	Score         float64   `json:"score"`
	Threshold     float64   `json:"threshold"`
	Pass          bool      `json:"pass"`
	Iterations    int       `json:"iterations"`
	Branch        string    `json:"branch"`
	FinishedAt    time.Time `json:"finished_at"`
	Findings      []Finding `json:"findings"`
}

func LoadReviewData(repoRoot string) (ReviewData, error) {
	path := filepath.Join(repoRoot, ".docs", "onboarding-loop", "latest-summary.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return ReviewData{}, err
	}
	var r RunResult
	if err := json.Unmarshal(raw, &r); err != nil {
		return ReviewData{}, err
	}
	return ReviewData{
		SchemaVersion: "v1",
		RunID:         r.RunID,
		Mode:          r.Mode,
		Score:         r.Judge.Score,
		Threshold:     r.Judge.Threshold,
		Pass:          r.Judge.Pass,
		Iterations:    r.Iterations,
		Branch:        r.Branch,
		FinishedAt:    r.FinishedAt,
		Findings:      r.Findings,
	}, nil
}
