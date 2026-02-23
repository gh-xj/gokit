# Onboarding Reinforcement Loop Design

Status: Approved (design)
Date: 2026-02-23
Owner: agentcli-go maintainers

## Goal

Build a hybrid local+CI onboarding reliability loop that:

1. mimics real user onboarding behavior in clean environments,
2. auto-fixes onboarding regressions on a dedicated branch,
3. records counter-intuitive UX issues in `.docs/`,
4. iterates until judge score is `>= 9.0/10`.

## Scope (v1)

- Execution mode: Hybrid (local fast loop + CI enforcement/verification).
- Autofix authority: auto-fix + auto-commit on dedicated branch.
- Scoring objective: balanced UX + quality + counter-intuitive penalties.

Out of scope (v1):

- fully autonomous merge to `main`,
- cross-repo orchestration beyond `agentcli-go` + tap update hook,
- model-based free-form patch generation (keep deterministic fix catalog first).

## Architecture

### Core components

1. `scenario`
- Declarative onboarding flows (5-minute flow + variants).
- Inputs: version channel (`main`/tag), temp dir root, environment toggles.
- Output: scenario execution spec.

2. `runner`
- Executes scenario in isolated temp directory.
- Captures per-step command, exit code, stdout/stderr, duration.
- Produces machine-readable run artifact.

3. `detector`
- Classifies findings:
  - `failure` (hard failure),
  - `counter_intuitive` (unexpected UX behavior),
  - `flaky` (non-deterministic behavior across reruns).

4. `fixer`
- Maps finding codes to deterministic fix strategies.
- Applies minimal patch and reruns affected scenarios.

5. `judge`
- Calculates score `/10`:
  - 40% onboarding UX,
  - 40% deterministic quality,
  - 20% counter-intuitive penalties.
- Produces pass/fail with explainable breakdown.

6. `reporter`
- Writes `.docs/onboarding-loop/` artifacts:
  - run summary,
  - finding registry,
  - score report,
  - iteration changelog.

7. `gitops`
- Uses dedicated branch: `autofix/onboarding-loop`.
- Commits only when score improves and verification is green.

## Data Flow

1. Load scenarios.
2. Run scenarios in temp environments.
3. Detect failures/UX friction.
4. Judge current score.
5. If score `< 9.0`, apply deterministic fix.
6. Re-run targeted scenario + `task ci`.
7. Write `.docs/` report artifacts.
8. Commit on `autofix/onboarding-loop` if improved and stable.
9. Stop when `>= 9.0` or iteration budget exhausted.

## Operational Interface

Planned tasks:

- `task loop:run`
- `task loop:judge`
- `task loop:autofix`
- `task loop:all`

Supporting controls:

- max iterations,
- dry-run mode,
- branch override,
- score threshold override (default `9.0`).

## Scoring Contract

### UX (40%)
- end-to-end pass rate,
- time-to-first-green,
- manual steps required.

### Quality (40%)
- `task ci` pass,
- contract fixtures/schemas pass,
- deterministic rerun consistency.

### Counter-intuitive penalties (20%)
- shell exits without clear root cause (for example `set -e` abrupt failures),
- ambiguous prerequisite failures,
- mismatch between docs and actual workflow behavior.

Gate: `score >= 9.0`.

## CI Model

- Local loop: fast iteration and autofix authoring.
- CI loop:
  - validates scenario runner and judge on PR/tag,
  - enforces score threshold,
  - blocks merge when below threshold.

## Risk Controls

- Deterministic fix catalog first (no unconstrained edits).
- Commit only after green verification.
- Keep reports append-only in `.docs/onboarding-loop/`.
- Explicit max iteration cap to avoid runaway loops.

## Deliverables

1. Harness loop module and task orchestration.
2. Scenario specs for canonical onboarding flows.
3. Judge scoring implementation + thresholds.
4. `.docs/onboarding-loop/` report contract.
5. Branch-based autofix workflow.

## Success Criteria

- Loop repeatedly catches regressions like scaffold formatting drift.
- Counter-intuitive issues are traceable in `.docs/`.
- Autofix branch produces minimal corrective commits.
- Judge score reaches `>= 9.0` consistently on stable mainline.
