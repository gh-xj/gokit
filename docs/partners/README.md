# Design Partner Program

This folder tracks external onboarding feedback for `agentcli-go`.

## Goal

Measure whether users can adopt `agentcli-go` quickly and reliably.

## Core Metrics

- `time_to_first_scaffold_success_min`
- `time_to_first_verify_pass_min`
- `doctor_iterations_before_green`
- `num_contract_related_failures`
- `num_clarification_questions`
- `overall_onboarding_score_1_10`

## Process

1. Create one file per partner using the template:
   - `YYYY-MM-DD-<partner>-onboarding.md`
2. Fill baseline environment and workflow notes.
3. Log timestamps and blockers during onboarding.
4. Summarize findings and action items.

## Success Thresholds (initial)

- 80%+ partners reach scaffold success in <= 10 min
- 70%+ partners reach first `task verify` pass in <= 20 min
- median `doctor_iterations_before_green` <= 2

## Weekly Review

Create/update `weekly-summary.md` with trend analysis and top 3 roadmap actions.
