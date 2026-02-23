package loopapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	harnessloop "github.com/gh-xj/agentcli-go/internal/harnessloop"
)

func Run(apiURL string, req RunRequest) (harnessloop.RunResult, error) {
	if strings.TrimSpace(apiURL) == "" {
		return harnessloop.RunResult{}, fmt.Errorf("api url is required")
	}
	body, err := json.Marshal(req)
	if err != nil {
		return harnessloop.RunResult{}, err
	}
	resp, err := http.Post(strings.TrimRight(apiURL, "/")+"/v1/loop/run", "application/json", bytes.NewReader(body))
	if err != nil {
		return harnessloop.RunResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var e map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&e)
		if msg, ok := e["error"]; ok && msg != "" {
			return harnessloop.RunResult{}, errors.New(msg)
		}
		return harnessloop.RunResult{}, fmt.Errorf("api request failed: %s", resp.Status)
	}
	var out harnessloop.RunResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return harnessloop.RunResult{}, err
	}
	return out, nil
}
