package harnessloop

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type plannerOutput struct {
	SchemaVersion string   `json:"schema_version"`
	Summary       string   `json:"summary"`
	FixTargets    []string `json:"fix_targets"`
}

type fixerOutput struct {
	SchemaVersion string   `json:"schema_version"`
	Applied       []string `json:"applied"`
	Notes         string   `json:"notes"`
}

type judgerOutput struct {
	SchemaVersion string    `json:"schema_version"`
	ExtraFindings []Finding `json:"extra_findings"`
	Notes         string    `json:"notes"`
}

type roleContext struct {
	RunID       string         `json:"run_id"`
	Mode        string         `json:"mode"`
	Iteration   int            `json:"iteration"`
	Threshold   float64        `json:"threshold"`
	Budget      int            `json:"budget"`
	Seed        int64          `json:"seed"`
	Scenario    ScenarioResult `json:"scenario"`
	Findings    []Finding      `json:"findings"`
	FixesSoFar  []string       `json:"fixes_so_far"`
	RepoRoot    string         `json:"repo_root"`
	ArtifactDir string         `json:"artifact_dir"`
}

func runCommittee(cfg Config, agentcliBin string, started time.Time, runID string) (RunResult, error) {
	roles, err := loadRoleConfig(cfg.RoleConfigPath)
	if err != nil {
		return RunResult{}, err
	}

	baseArtifacts := filepath.Join(cfg.RepoRoot, ".docs", "onboarding-loop", "runs", runID)
	if err := os.MkdirAll(baseArtifacts, 0755); err != nil {
		return RunResult{}, err
	}

	result := RunResult{
		SchemaVersion: "v1",
		StartedAt:     started,
		Branch:        CurrentBranch(cfg.RepoRoot),
		Mode:          cfg.Mode,
		RunID:         runID,
		Committee: &CommitteeMeta{
			Planner: RoleExecution{Strategy: strategyOrBuiltin(roles.Planner)},
			Fixer:   RoleExecution{Strategy: strategyOrBuiltin(roles.Fixer)},
			Judger:  RoleExecution{Strategy: strategyOrBuiltin(roles.Judger), Independent: true},
		},
	}

	for i := 1; i <= cfg.MaxIterations; i++ {
		iterArtifacts := filepath.Join(baseArtifacts, fmt.Sprintf("iter-%02d", i))
		if err := os.MkdirAll(iterArtifacts, 0755); err != nil {
			return result, err
		}

		sr, findings, err := runScenarioAndFindings(agentcliBin, cfg.RepoRoot)
		if err != nil {
			return result, err
		}
		ctx := roleContext{
			RunID:       runID,
			Mode:        cfg.Mode,
			Iteration:   i,
			Threshold:   cfg.Threshold,
			Budget:      cfg.Budget,
			Seed:        cfg.Seed,
			Scenario:    sr,
			Findings:    findings,
			FixesSoFar:  append([]string{}, result.FixesApplied...),
			RepoRoot:    cfg.RepoRoot,
			ArtifactDir: iterArtifacts,
		}

		plan, plannerExec, err := runPlannerRole(cfg.RepoRoot, iterArtifacts, roles.Planner, ctx)
		if err != nil {
			return result, err
		}
		result.Committee.Planner = plannerExec

		fixes, fixerExec, err := runFixerRole(cfg.RepoRoot, iterArtifacts, roles.Fixer, ctx, findings, plan)
		if err != nil {
			return result, err
		}
		result.Committee.Fixer = fixerExec
		result.FixesApplied = append(result.FixesApplied, fixes...)

		postScenario, postFindings, err := runScenarioAndFindings(agentcliBin, cfg.RepoRoot)
		if err != nil {
			return result, err
		}
		jCtx := ctx
		jCtx.Scenario = postScenario
		jCtx.Findings = postFindings
		judgeOut, judgerExec, err := runJudgerRole(cfg.RepoRoot, iterArtifacts, roles.Judger, jCtx)
		if err != nil {
			return result, err
		}
		result.Committee.Judger = judgerExec

		allFindings := append([]Finding{}, postFindings...)
		allFindings = append(allFindings, judgeOut.ExtraFindings...)
		score := Judge(postScenario, allFindings, cfg.Threshold)
		score.PlannerScore = scorePlanner(plan, findings)
		score.FixerScore = scoreFixer(findings, allFindings, fixes)
		score.JudgerScore = scoreJudger(judgeOut, result.Committee.Judger)

		result.Scenario = postScenario
		result.Findings = allFindings
		result.Judge = score
		result.Iterations = i
		result.FinishedAt = time.Now().UTC()

		if score.Pass || !cfg.AutoFix || len(fixes) == 0 {
			break
		}
	}

	if result.Iterations == 0 {
		result.Iterations = 1
		result.FinishedAt = time.Now().UTC()
	}

	return result, writeJSON(filepath.Join(baseArtifacts, "final-report.json"), result)
}

func runScenarioAndFindings(agentcliBin, repoRoot string) (ScenarioResult, []Finding, error) {
	scenario := DefaultOnboardingScenario(agentcliBin)
	sr, err := RunScenario(scenario)
	if err != nil {
		return ScenarioResult{}, nil, err
	}
	findings := DetectFindings(sr)
	findings = append(findings, CheckOnboardingInstallReadiness(repoRoot)...)
	return sr, findings, nil
}

func strategyOrBuiltin(spec RoleSpec) string {
	if strings.TrimSpace(spec.Strategy) != "" {
		return strings.TrimSpace(spec.Strategy)
	}
	if strings.TrimSpace(spec.Command) != "" {
		return "external"
	}
	return "builtin"
}

func loadRoleConfig(path string) (RoleConfig, error) {
	if strings.TrimSpace(path) == "" {
		return RoleConfig{}, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return RoleConfig{}, fmt.Errorf("read role config: %w", err)
	}
	var cfg RoleConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return RoleConfig{}, fmt.Errorf("parse role config: %w", err)
	}
	return cfg, nil
}

func scorePlanner(plan plannerOutput, findings []Finding) float64 {
	if len(findings) == 0 {
		return 5.0
	}
	if len(plan.FixTargets) == 0 {
		return 1.0
	}
	targets := map[string]bool{}
	for _, t := range plan.FixTargets {
		targets[strings.TrimSpace(t)] = true
	}
	matched := 0
	for _, f := range findings {
		if targets[f.Code] {
			matched++
		}
	}
	ratio := float64(matched) / float64(len(findings))
	return 1.0 + ratio*4.0
}

func scoreFixer(before, after []Finding, fixes []string) float64 {
	if len(before) == 0 {
		return 5.0
	}
	if len(fixes) == 0 {
		return 1.0
	}
	reduction := len(before) - len(after)
	if reduction < 0 {
		return 1.0
	}
	ratio := float64(reduction) / float64(len(before))
	return 1.0 + ratio*4.0
}

func scoreJudger(out judgerOutput, exec RoleExecution) float64 {
	if !exec.Independent {
		return 2.5
	}
	if strings.TrimSpace(out.Notes) == "" {
		return 2.0
	}
	return 5.0
}
