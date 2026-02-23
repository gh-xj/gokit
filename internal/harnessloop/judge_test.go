package harnessloop

import "testing"

func TestJudgePassesHealthyRun(t *testing.T) {
	r := ScenarioResult{OK: true}
	s := Judge(r, nil, 9.0)
	if !s.Pass {
		t.Fatalf("expected pass at threshold 9.0, got score %.2f", s.Score)
	}
	if s.Score != 10.0 {
		t.Fatalf("unexpected score: %.2f", s.Score)
	}
}

func TestJudgePenalizesFindings(t *testing.T) {
	r := ScenarioResult{OK: false}
	findings := []Finding{{Code: "step_failed"}, {Code: "counter_intuitive_abort"}}
	s := Judge(r, findings, 1.0)
	if s.Score >= 1.0 {
		t.Fatalf("expected low score, got %.2f", s.Score)
	}
}
