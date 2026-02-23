package loopapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	harnessloop "github.com/gh-xj/agentcli-go/internal/harnessloop"
)

func TestRunClientSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/loop/run" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(harnessloop.RunResult{
			SchemaVersion: "v1",
			Judge: harnessloop.JudgeScore{
				Score: 9.5,
				Pass:  true,
			},
		})
	}))
	defer ts.Close()

	out, err := Run(ts.URL, RunRequest{Action: "judge"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.Judge.Pass || out.Judge.Score < 9 {
		t.Fatalf("unexpected score: %+v", out.Judge)
	}
}
