package harnessloop

import "testing"

func TestDetectFindingsFromFailedStep(t *testing.T) {
	r := ScenarioResult{
		OK: false,
		Steps: []StepResult{{
			Name:         "verify",
			ExitCode:     1,
			CombinedTail: "task: Failed to run task \"verify\": task: Failed to run task \"build\": task: Failed to run task \"fmt:check\": exit status 1",
		}},
	}
	f := DetectFindings(r)
	if len(f) == 0 {
		t.Fatal("expected findings")
	}
}
