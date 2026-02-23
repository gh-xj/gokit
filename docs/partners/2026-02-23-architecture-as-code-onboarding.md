# Partner Onboarding Report

- Date: 2026-02-23
- Partner: architecture-as-code (internal design partner)
- Repo/Use case: architecture check and generation utility CLI
- OS: macOS
- Go version: go1.25.x
- Agent setup: terminal coding agent with contract verification

## Timeline

- Start time: 2026-02-23 (local)
- First `agentcli new` success time: under 1 minute
- First `task verify` success time: about 1 minute total from start

## Metrics

- time_to_first_scaffold_success_min: 1
- time_to_first_verify_pass_min: 1
- doctor_iterations_before_green: 1
- num_contract_related_failures: 0
- num_clarification_questions: 0
- overall_onboarding_score_1_10: 9

## Friction Log

1. No contract-level failures.
2. Initial command naming conventions needed brief clarification.

## What Worked Well

- Health checks and verify flow were clear and agent-friendly.
- Contract boundaries reduced ambiguity for generated edits.

## Improvement Requests

- Add docs section mapping common architecture-use-case commands.

## Action Items for agentcli-go

1. Add advanced command naming guidance in docs.
2. Keep partner tracker trends visible in weekly summaries.
