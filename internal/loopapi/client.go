package loopapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	harnessloop "github.com/gh-xj/agentcli-go/internal/harnessloop"
	harness "github.com/gh-xj/agentcli-go/tools/harness"
)

func Run(apiURL string, req RunRequest) (harnessloop.RunResult, error) {
	summary, err := RunSummary(apiURL, req)
	if err != nil {
		return harnessloop.RunResult{}, err
	}
	var out harnessloop.RunResult
	if summary.Data != nil {
		raw, err := json.Marshal(summary.Data)
		if err != nil {
			return harnessloop.RunResult{}, err
		}
		if err := json.Unmarshal(raw, &out); err != nil {
			return harnessloop.RunResult{}, err
		}
	}
	if out.SchemaVersion != "" {
		return out, nil
	}
	if len(summary.Failures) > 0 {
		return harnessloop.RunResult{}, errors.New(summary.Failures[0].Message)
	}
	if summary.Status == harness.StatusFail {
		return harnessloop.RunResult{}, errors.New("loop api returned failure summary")
	}
	return harnessloop.RunResult{}, errors.New("loop api response missing run result")
}

func RunSummary(apiURL string, req RunRequest) (harness.CommandSummary, error) {
	if strings.TrimSpace(apiURL) == "" {
		return harness.CommandSummary{}, fmt.Errorf("api url is required")
	}
	body, err := json.Marshal(req)
	if err != nil {
		return harness.CommandSummary{}, err
	}
	resp, err := http.Post(strings.TrimRight(apiURL, "/")+"/v1/loop/run", "application/json", bytes.NewReader(body))
	if err != nil {
		return harness.CommandSummary{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var e map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&e)
		if msg, ok := e["error"]; ok && msg != "" {
			return harness.CommandSummary{}, errors.New(msg)
		}
		return harness.CommandSummary{}, fmt.Errorf("api request failed: %s", resp.Status)
	}
	var summary harness.CommandSummary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return harness.CommandSummary{}, err
	}
	return summary, nil
}
