# Onboarding Reinforcement Loop Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a hybrid local+CI onboarding reinforcement loop that auto-fixes regressions, tracks UX friction in `.docs/`, and gates on judge score `>= 9.0`.

**Architecture:** Implement a deterministic harness loop (`scenario -> runner -> detector -> fixer -> judge -> reporter -> gitops`) with task entrypoints and CI verification. Keep v1 constrained to a curated fix catalog and explicit scoring contract.

**Tech Stack:** Go 1.25+, Taskfile, GitHub Actions, existing `agentcli-go` test/contract harness.

---

### Task 1: Define report and scenario contracts

**Files:**
- Create: `internal/harnessloop/types.go`
- Create: `.docs/onboarding-loop/README.md`
- Test: `internal/harnessloop/types_test.go`

**Step 1: Write failing tests for JSON contract stability**
- Add tests asserting required fields for:
  - scenario spec,
  - run result,
  - finding,
  - judge score report.

**Step 2: Run focused test to confirm failure**
- Run: `go test ./internal/harnessloop -run TestContracts -v`
- Expected: FAIL (missing types/fields).

**Step 3: Add minimal contract structs + marshal helpers**
- Implement typed structs with stable JSON tags.

**Step 4: Re-run focused tests**
- Run: `go test ./internal/harnessloop -run TestContracts -v`
- Expected: PASS.

**Step 5: Commit**
- `git commit -m "feat: add onboarding loop type contracts"`

### Task 2: Implement scenario runner for onboarding scripts

**Files:**
- Create: `internal/harnessloop/runner.go`
- Create: `internal/harnessloop/runner_test.go`
- Create: `internal/harnessloop/scenarios.go`

**Step 1: Write failing tests for isolated execution**
- Verify:
  - temp dir isolation,
  - per-step timing capture,
  - stdout/stderr capture,
  - non-zero exit propagation.

**Step 2: Run tests to confirm failure**
- `go test ./internal/harnessloop -run TestRunner -v`

**Step 3: Implement minimal runner**
- Use deterministic command execution and structured step results.

**Step 4: Re-run tests**
- `go test ./internal/harnessloop -run TestRunner -v`

**Step 5: Commit**
- `git commit -m "feat: add deterministic onboarding scenario runner"`

### Task 3: Add detector and finding classification

**Files:**
- Create: `internal/harnessloop/detector.go`
- Create: `internal/harnessloop/detector_test.go`

**Step 1: Write failing tests for finding classification**
- Cases:
  - hard failure,
  - counter-intuitive behavior (abrupt set-e stop),
  - flaky rerun mismatch.

**Step 2: Run failing tests**
- `go test ./internal/harnessloop -run TestDetector -v`

**Step 3: Implement deterministic classification rules**
- Map regex/signatures to finding codes.

**Step 4: Re-run tests**
- `go test ./internal/harnessloop -run TestDetector -v`

**Step 5: Commit**
- `git commit -m "feat: add onboarding finding detector"`

### Task 4: Add judge scoring model (balanced)

**Files:**
- Create: `internal/harnessloop/judge.go`
- Create: `internal/harnessloop/judge_test.go`

**Step 1: Write failing score tests**
- Validate weighted score:
  - UX 40,
  - quality 40,
  - penalties 20.
- Validate threshold logic (`>= 9.0` pass).

**Step 2: Run failing tests**
- `go test ./internal/harnessloop -run TestJudge -v`

**Step 3: Implement scoring engine**
- Produce explainable score breakdown and verdict.

**Step 4: Re-run tests**
- `go test ./internal/harnessloop -run TestJudge -v`

**Step 5: Commit**
- `git commit -m "feat: add onboarding judge scoring"`

### Task 5: Add report writer to `.docs/onboarding-loop`

**Files:**
- Create: `internal/harnessloop/reporter.go`
- Create: `internal/harnessloop/reporter_test.go`
- Create: `.docs/onboarding-loop/.gitkeep`

**Step 1: Write failing tests for artifact output**
- Ensure files created:
  - `latest-summary.json`,
  - timestamped markdown report,
  - findings registry update.

**Step 2: Run failing tests**
- `go test ./internal/harnessloop -run TestReporter -v`

**Step 3: Implement reporter**
- Stable deterministic filenames and append behavior.

**Step 4: Re-run tests**
- `go test ./internal/harnessloop -run TestReporter -v`

**Step 5: Commit**
- `git commit -m "feat: add onboarding loop reporting artifacts"`

### Task 6: Add fixer catalog + branch commit workflow

**Files:**
- Create: `internal/harnessloop/fixer.go`
- Create: `internal/harnessloop/fixer_test.go`
- Create: `internal/harnessloop/gitops.go`
- Create: `internal/harnessloop/gitops_test.go`

**Step 1: Write failing tests**
- Verify:
  - known finding -> known fix strategy,
  - no-op for unknown finding,
  - branch creation/switch to `autofix/onboarding-loop`,
  - commit only after green.

**Step 2: Run failing tests**
- `go test ./internal/harnessloop -run 'TestFixer|TestGitOps' -v`

**Step 3: Implement minimal deterministic fix catalog + gitops**
- Include formatting drift, doc-command mismatch, missing prerequisite hints.

**Step 4: Re-run tests**
- `go test ./internal/harnessloop -run 'TestFixer|TestGitOps' -v`

**Step 5: Commit**
- `git commit -m "feat: add onboarding autofix catalog and gitops"`

### Task 7: Expose loop CLI/task entrypoints

**Files:**
- Create: `cmd/agentcli/loop.go`
- Modify: `cmd/agentcli/main.go`
- Modify: `cmd/agentcli/main_test.go`
- Modify: `Taskfile.yml`

**Step 1: Write failing command tests**
- `agentcli loop run|judge|autofix|all` behavior and exit codes.

**Step 2: Run failing tests**
- `go test ./cmd/agentcli -run TestRunLoop -v`

**Step 3: Implement command wiring + task wrappers**
- Add:
  - `task loop:run`
  - `task loop:judge`
  - `task loop:autofix`
  - `task loop:all`

**Step 4: Re-run command tests and repo CI**
- `go test ./cmd/agentcli -run TestRunLoop -v`
- `task ci`

**Step 5: Commit**
- `git commit -m "feat: add onboarding loop commands and tasks"`

### Task 8: Add CI enforcement workflow

**Files:**
- Create: `.github/workflows/onboarding-loop.yml`
- Modify: `README.md`

**Step 1: Add failing workflow smoke check (local via act optional)**
- Ensure workflow references valid commands/files.

**Step 2: Implement workflow**
- Run judge on PR and fail below threshold.
- Persist `.docs` artifacts as workflow artifacts.

**Step 3: Validate**
- `task ci`
- Dry-run workflow syntax check (or push branch and inspect Actions).

**Step 4: Commit**
- `git commit -m "ci: add onboarding loop judge workflow"`

### Task 9: End-to-end validation and baseline score

**Files:**
- Modify: `.docs/onboarding-loop/baseline.md`
- Modify: `docs/site/ai-agent-playbook.md`

**Step 1: Run full loop**
- `task loop:all`

**Step 2: Verify outcome**
- Judge score `>= 9.0`
- No unstable rerun behavior across 3 consecutive runs.

**Step 3: Document baseline + known limits**
- Capture score breakdown and remaining friction.

**Step 4: Commit**
- `git commit -m "docs: record onboarding loop baseline and score"`

### Task 10: Final verification gate

**Files:**
- Modify (if needed): `README.md`, `Taskfile.yml`, `.github/workflows/*`

**Step 1: Run full verification**
- `task ci`
- `task release:build VERSION=v0.2.2-rc1` (or next candidate)

**Step 2: Confirm no unexpected diffs**
- `git status --short`

**Step 3: Merge readiness report**
- Summarize score, findings trend, and guardrails.

**Step 4: Commit**
- `git commit -m "chore: finalize onboarding reinforcement loop v1"`

---

Plan complete and saved to `docs/plans/2026-02-23-onboarding-reinforcement-loop-implementation-plan.md`.

Two execution options:

1. Subagent-Driven (this session) - dispatch fresh subagent per task with reviews between tasks.
2. Parallel Session (separate) - execute in dedicated worktree using executing-plans checkpoints.

Which approach?
