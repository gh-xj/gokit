package loopapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	harnessloop "github.com/gh-xj/agentcli-go/internal/harnessloop"
	harness "github.com/gh-xj/agentcli-go/tools/harness"
)

func TestRunClientSuccess(t *testing.T) {
	var got RunRequest
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/loop/run" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(harness.CommandSummary{
			SchemaVersion: harness.SummarySchemaVersion,
			Command:       "loop judge",
			Status:        harness.StatusOK,
			Data: harnessloop.RunResult{
				SchemaVersion: "v1",
				Judge: harnessloop.JudgeScore{
					Score: 9.5,
					Pass:  true,
				},
			},
		})
	}))
	defer ts.Close()

	out, err := Run(ts.URL, RunRequest{
		Action:           "judge",
		RepoRoot:         "/tmp/repo",
		Threshold:        8.5,
		MaxIterations:    4,
		Branch:           "autofix/test",
		Mode:             "committee",
		RoleConfig:       "configs/skill-quality.roles.json",
		VerboseArtifacts: true,
		Seed:             42,
		Budget:           2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.Judge.Pass || out.Judge.Score < 9 {
		t.Fatalf("unexpected score: %+v", out.Judge)
	}
	if got.Action != "judge" || got.RepoRoot != "/tmp/repo" || got.Threshold != 8.5 || got.MaxIterations != 4 || got.Branch != "autofix/test" || got.Mode != "committee" || got.RoleConfig != "configs/skill-quality.roles.json" || !got.VerboseArtifacts || got.Seed != 42 || got.Budget != 2 {
		t.Fatalf("unexpected request payload: %+v", got)
	}
}

func TestRunSummarySuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(harness.CommandSummary{
			SchemaVersion: harness.SummarySchemaVersion,
			Command:       "loop run",
			Status:        harness.StatusOK,
			Data:          map[string]any{"ok": true},
		})
	}))
	defer ts.Close()

	summary, err := RunSummary(ts.URL, RunRequest{Action: "run"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Command != "loop run" || summary.Status != harness.StatusOK {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

func TestRunClientSummaryFailureWithoutData(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(harness.CommandSummary{
			SchemaVersion: harness.SummarySchemaVersion,
			Command:       "loop run",
			Status:        harness.StatusFail,
			Failures: []harness.Failure{
				{Code: string(harness.CodeExecution), Message: "loop run failed"},
			},
		})
	}))
	defer ts.Close()

	_, err := Run(ts.URL, RunRequest{Action: "run"})
	if err == nil {
		t.Fatal("expected error")
	}
}
