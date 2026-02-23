package harnessloop

import (
	"encoding/json"
	"testing"
)

func TestRunResultJSONContract(t *testing.T) {
	r := RunResult{SchemaVersion: "v1", Scenario: ScenarioResult{Name: "default"}, Judge: JudgeScore{Score: 8.0}}
	out, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("empty json")
	}
}
