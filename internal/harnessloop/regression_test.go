package harnessloop

import (
	"path/filepath"
	"testing"
	"time"
)

func TestBuildBehaviorSnapshot(t *testing.T) {
	run := RunResult{
		Mode:       "committee",
		Iterations: 2,
		Scenario: ScenarioResult{
			Name: "default-onboarding",
			OK:   true,
			Steps: []StepResult{
				{Name: "doctor", ExitCode: 0, DurationMs: 123},
				{Name: "verify", ExitCode: 0, DurationMs: 456},
			},
		},
		Findings: []Finding{
			{Code: "b", Severity: "low", Source: "x"},
			{Code: "a", Severity: "high", Source: "y"},
		},
		Judge: JudgeScore{
			Threshold:            9.0,
			Score:                10.0,
			Pass:                 true,
			HardFailures:         0,
			CounterIntuitiveFind: 1,
		},
		Committee: &CommitteeMeta{
			Planner: RoleExecution{Strategy: "builtin"},
			Fixer:   RoleExecution{Strategy: "external"},
			Judger:  RoleExecution{Strategy: "builtin"},
		},
	}

	snapshot := BuildBehaviorSnapshot(run)
	if snapshot.Mode != "committee" || snapshot.Iterations != 2 || !snapshot.Scenario.OK {
		t.Fatalf("unexpected snapshot header: %+v", snapshot)
	}
	if len(snapshot.Scenario.Steps) != 2 || snapshot.Scenario.Steps[0].Name != "doctor" {
		t.Fatalf("unexpected step snapshot: %+v", snapshot.Scenario.Steps)
	}
	if len(snapshot.Findings) != 2 || snapshot.Findings[0].Code != "a" || snapshot.Findings[1].Code != "b" {
		t.Fatalf("findings should be sorted and reduced: %+v", snapshot.Findings)
	}
	if snapshot.Committee == nil || snapshot.Committee.FixerStrategy != "external" {
		t.Fatalf("unexpected committee snapshot: %+v", snapshot.Committee)
	}
}

func TestCompareBehaviorSnapshot(t *testing.T) {
	expected := BehaviorSnapshot{
		Mode:       "committee",
		Iterations: 1,
		Scenario: BehaviorScenarioSnapshot{
			Name: "s",
			OK:   true,
			Steps: []BehaviorStepSnapshot{
				{Name: "doctor", ExitCode: 0},
			},
		},
		Findings: []BehaviorFindingSnapshot{
			{Code: "f1", Severity: "high", Source: "README.md"},
		},
		Judge: BehaviorJudgeSnapshot{
			Threshold:               9.0,
			Score:                   10.0,
			Pass:                    true,
			HardFailures:            0,
			CounterIntuitiveFinding: 0,
		},
	}
	actual := expected
	drifts := CompareBehaviorSnapshot(expected, actual)
	if len(drifts) != 0 {
		t.Fatalf("expected no drifts, got %+v", drifts)
	}

	actual.Scenario.OK = false
	actual.Judge.Score = 8.5
	drifts = CompareBehaviorSnapshot(expected, actual)
	if len(drifts) == 0 {
		t.Fatal("expected drifts")
	}
}

func TestRegressionBaselineRoundTrip(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "testdata", "regression", "baseline.json")
	baseline := RegressionBaseline{
		SchemaVersion: "v1",
		Kind:          "loop_behavior",
		Profile:       "quality",
		GeneratedAt:   time.Now().UTC(),
		Snapshot: BehaviorSnapshot{
			Mode:       "committee",
			Iterations: 1,
		},
	}
	if err := WriteRegressionBaseline(path, baseline); err != nil {
		t.Fatalf("write baseline: %v", err)
	}
	got, err := ReadRegressionBaseline(path)
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}
	if got.SchemaVersion != "v1" || got.Kind != "loop_behavior" || got.Profile != "quality" || got.Snapshot.Mode != "committee" {
		t.Fatalf("unexpected round-trip baseline: %+v", got)
	}
}
