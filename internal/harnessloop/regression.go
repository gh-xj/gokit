package harnessloop

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	regressionSchemaVersion = "v1"
	regressionBaselineKind  = "loop_behavior"
)

type BehaviorSnapshot struct {
	Mode       string                     `json:"mode"`
	Iterations int                        `json:"iterations"`
	Scenario   BehaviorScenarioSnapshot   `json:"scenario"`
	Findings   []BehaviorFindingSnapshot  `json:"findings"`
	Judge      BehaviorJudgeSnapshot      `json:"judge"`
	Committee  *BehaviorCommitteeSnapshot `json:"committee,omitempty"`
}

type BehaviorScenarioSnapshot struct {
	Name  string                 `json:"name"`
	OK    bool                   `json:"ok"`
	Steps []BehaviorStepSnapshot `json:"steps"`
}

type BehaviorStepSnapshot struct {
	Name     string `json:"name"`
	ExitCode int    `json:"exit_code"`
}

type BehaviorFindingSnapshot struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Source   string `json:"source"`
}

type BehaviorJudgeSnapshot struct {
	Threshold               float64 `json:"threshold"`
	Score                   float64 `json:"score"`
	Pass                    bool    `json:"pass"`
	HardFailures            int     `json:"hard_failures"`
	CounterIntuitiveFinding int     `json:"counter_intuitive_findings"`
}

type BehaviorCommitteeSnapshot struct {
	PlannerStrategy string `json:"planner_strategy"`
	FixerStrategy   string `json:"fixer_strategy"`
	JudgerStrategy  string `json:"judger_strategy"`
}

type RegressionBaseline struct {
	SchemaVersion string           `json:"schema_version"`
	Kind          string           `json:"kind"`
	Profile       string           `json:"profile"`
	GeneratedAt   time.Time        `json:"generated_at"`
	Snapshot      BehaviorSnapshot `json:"snapshot"`
}

type RegressionDrift struct {
	Path     string `json:"path"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
}

func BuildBehaviorSnapshot(result RunResult) BehaviorSnapshot {
	steps := make([]BehaviorStepSnapshot, 0, len(result.Scenario.Steps))
	for _, step := range result.Scenario.Steps {
		steps = append(steps, BehaviorStepSnapshot{
			Name:     step.Name,
			ExitCode: step.ExitCode,
		})
	}

	findings := make([]BehaviorFindingSnapshot, 0, len(result.Findings))
	for _, finding := range result.Findings {
		findings = append(findings, BehaviorFindingSnapshot{
			Code:     finding.Code,
			Severity: finding.Severity,
			Source:   finding.Source,
		})
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Code != findings[j].Code {
			return findings[i].Code < findings[j].Code
		}
		if findings[i].Severity != findings[j].Severity {
			return findings[i].Severity < findings[j].Severity
		}
		return findings[i].Source < findings[j].Source
	})

	var committee *BehaviorCommitteeSnapshot
	if result.Committee != nil {
		committee = &BehaviorCommitteeSnapshot{
			PlannerStrategy: result.Committee.Planner.Strategy,
			FixerStrategy:   result.Committee.Fixer.Strategy,
			JudgerStrategy:  result.Committee.Judger.Strategy,
		}
	}

	return BehaviorSnapshot{
		Mode:       result.Mode,
		Iterations: result.Iterations,
		Scenario: BehaviorScenarioSnapshot{
			Name:  result.Scenario.Name,
			OK:    result.Scenario.OK,
			Steps: steps,
		},
		Findings: findings,
		Judge: BehaviorJudgeSnapshot{
			Threshold:               result.Judge.Threshold,
			Score:                   result.Judge.Score,
			Pass:                    result.Judge.Pass,
			HardFailures:            result.Judge.HardFailures,
			CounterIntuitiveFinding: result.Judge.CounterIntuitiveFind,
		},
		Committee: committee,
	}
}

func CompareBehaviorSnapshot(expected, actual BehaviorSnapshot) []RegressionDrift {
	drifts := make([]RegressionDrift, 0)

	add := func(path string, want, got any) {
		if fmt.Sprintf("%v", want) == fmt.Sprintf("%v", got) {
			return
		}
		drifts = append(drifts, RegressionDrift{
			Path:     path,
			Expected: fmt.Sprintf("%v", want),
			Actual:   fmt.Sprintf("%v", got),
		})
	}

	add("mode", expected.Mode, actual.Mode)
	add("iterations", expected.Iterations, actual.Iterations)
	add("scenario.name", expected.Scenario.Name, actual.Scenario.Name)
	add("scenario.ok", expected.Scenario.OK, actual.Scenario.OK)

	add("scenario.steps.length", len(expected.Scenario.Steps), len(actual.Scenario.Steps))
	stepsLen := len(expected.Scenario.Steps)
	if len(actual.Scenario.Steps) < stepsLen {
		stepsLen = len(actual.Scenario.Steps)
	}
	for i := 0; i < stepsLen; i++ {
		add(fmt.Sprintf("scenario.steps[%d].name", i), expected.Scenario.Steps[i].Name, actual.Scenario.Steps[i].Name)
		add(fmt.Sprintf("scenario.steps[%d].exit_code", i), expected.Scenario.Steps[i].ExitCode, actual.Scenario.Steps[i].ExitCode)
	}

	add("findings.length", len(expected.Findings), len(actual.Findings))
	findingsLen := len(expected.Findings)
	if len(actual.Findings) < findingsLen {
		findingsLen = len(actual.Findings)
	}
	for i := 0; i < findingsLen; i++ {
		add(fmt.Sprintf("findings[%d].code", i), expected.Findings[i].Code, actual.Findings[i].Code)
		add(fmt.Sprintf("findings[%d].severity", i), expected.Findings[i].Severity, actual.Findings[i].Severity)
		add(fmt.Sprintf("findings[%d].source", i), expected.Findings[i].Source, actual.Findings[i].Source)
	}

	add("judge.pass", expected.Judge.Pass, actual.Judge.Pass)
	add("judge.hard_failures", expected.Judge.HardFailures, actual.Judge.HardFailures)
	add("judge.counter_intuitive_findings", expected.Judge.CounterIntuitiveFinding, actual.Judge.CounterIntuitiveFinding)
	add("judge.threshold", round3(expected.Judge.Threshold), round3(actual.Judge.Threshold))
	add("judge.score", round3(expected.Judge.Score), round3(actual.Judge.Score))

	if expected.Committee == nil || actual.Committee == nil {
		add("committee.present", expected.Committee != nil, actual.Committee != nil)
		return drifts
	}
	add("committee.planner_strategy", expected.Committee.PlannerStrategy, actual.Committee.PlannerStrategy)
	add("committee.fixer_strategy", expected.Committee.FixerStrategy, actual.Committee.FixerStrategy)
	add("committee.judger_strategy", expected.Committee.JudgerStrategy, actual.Committee.JudgerStrategy)

	return drifts
}

func ReadRegressionBaseline(path string) (RegressionBaseline, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return RegressionBaseline{}, err
	}
	var baseline RegressionBaseline
	if err := json.Unmarshal(raw, &baseline); err != nil {
		return RegressionBaseline{}, fmt.Errorf("parse regression baseline %s: %w", path, err)
	}
	if baseline.SchemaVersion != regressionSchemaVersion {
		return RegressionBaseline{}, fmt.Errorf("invalid baseline schema_version: %q", baseline.SchemaVersion)
	}
	if baseline.Kind != "" && baseline.Kind != regressionBaselineKind {
		return RegressionBaseline{}, fmt.Errorf("invalid baseline kind: %q", baseline.Kind)
	}
	return baseline, nil
}

func WriteRegressionBaseline(path string, baseline RegressionBaseline) error {
	if baseline.SchemaVersion == "" {
		baseline.SchemaVersion = regressionSchemaVersion
	}
	if baseline.Kind == "" {
		baseline.Kind = regressionBaselineKind
	}
	if baseline.GeneratedAt.IsZero() {
		baseline.GeneratedAt = time.Now().UTC()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return writeJSON(path, baseline)
}

func round3(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return v
	}
	return math.Round(v*1000) / 1000
}
