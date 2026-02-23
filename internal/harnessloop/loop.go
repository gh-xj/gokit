package harnessloop

import (
	"fmt"
	"time"
)

type Config struct {
	RepoRoot      string
	Threshold     float64
	MaxIterations int
	AutoFix       bool
	AutoCommit    bool
	Branch        string
}

func RunLoop(cfg Config) (RunResult, error) {
	if cfg.Threshold <= 0 {
		cfg.Threshold = 9.0
	}
	if cfg.MaxIterations <= 0 {
		cfg.MaxIterations = 3
	}
	if cfg.Branch == "" {
		cfg.Branch = "autofix/onboarding-loop"
	}

	started := time.Now().UTC()
	agentcliBin, err := BuildLocalAgentCLIBinary(cfg.RepoRoot)
	if err != nil {
		return RunResult{}, err
	}

	var best RunResult
	best.Judge.Score = -1
	iterations := 0
	for ; iterations < cfg.MaxIterations; iterations++ {
		scenario := DefaultOnboardingScenario(agentcliBin)
		sr, err := RunScenario(scenario)
		if err != nil {
			return RunResult{}, err
		}
		findings := DetectFindings(sr)
		judge := Judge(sr, findings, cfg.Threshold)
		run := RunResult{
			SchemaVersion: "v1",
			StartedAt:     started,
			FinishedAt:    time.Now().UTC(),
			Scenario:      sr,
			Findings:      findings,
			Judge:         judge,
			Iterations:    iterations + 1,
			Branch:        CurrentBranch(cfg.RepoRoot),
		}
		if judge.Score > best.Judge.Score {
			best = run
		}
		if judge.Pass || !cfg.AutoFix {
			break
		}
		applied, err := ApplyFixes(cfg.RepoRoot, findings)
		if err != nil {
			return run, err
		}
		run.FixesApplied = append(run.FixesApplied, applied...)
		best = run
		if len(applied) == 0 {
			break
		}
	}
	best.Iterations = iterations + 1
	best.FinishedAt = time.Now().UTC()

	if err := WriteReports(cfg.RepoRoot, best); err != nil {
		return best, err
	}

	if cfg.AutoCommit {
		if err := EnsureBranch(cfg.RepoRoot, cfg.Branch); err != nil {
			return best, err
		}
		committed, err := CommitIfDirty(cfg.RepoRoot, fmt.Sprintf("chore: onboarding loop autofix score %.2f", best.Judge.Score))
		if err != nil {
			return best, err
		}
		if committed {
			best.Branch = cfg.Branch
		}
	}
	return best, nil
}
