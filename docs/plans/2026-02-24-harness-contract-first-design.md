# Harness Contract-First Redesign (Aggressive Cutover) Design

## Context

This repository currently exposes harness workflows through `cmd/agentcli/loop.go` and `internal/harnessloop/*`.
Behavior has improved, but output, failure semantics, and command metadata are still split across command paths.
The proposed upgrade is to build a new machine-first harness core under `tools/harness` and cut over aggressively with no backward compatibility layer.

## Objective

Build a contract-first harness runtime that is:

- machine-consumable by default (`json`/`ndjson`/typed failures),
- safe for automation (`dry-run`/`explain`),
- deterministic in CI (summary on both success and failure),
- self-discoverable (`capabilities` command).

## Decision Summary

Chosen direction: **most aggressive and elegant**.

1. Build new canonical core in `tools/harness`.
2. Convert `cmd/agentcli/loop.go` into a thin adapter immediately.
3. Remove legacy runtime behavior from primary execution path (no compatibility window).

## Section 1: Target Architecture (Approved)

Primary source of truth becomes `tools/harness`:

- `tools/harness/contract.go`
  - common command summary schema and shared status/check/failure models.
- `tools/harness/errors.go`
  - typed error envelope and deterministic exit-code mapping.
- `tools/harness/runner.go`
  - shared command wrapper: timing, execution lifecycle, summary emission, and output rendering.
- `tools/harness/capabilities.go`
  - runtime discovery model for supported commands/flags/artifacts/schema.
- `tools/harness/commands/*`
  - command implementations: `doctor`, `quality`, `lean`, `regression`, and lab utilities.

`cmd/agentcli/loop.go` becomes adapter-only:

- parse CLI args,
- resolve command + global flags,
- call `tools/harness`,
- render `text|json|ndjson` via unified contract renderer.

## Section 2: Unified Command Contract (Approved)

Global flags for all harness commands:

- `--format text|json|ndjson` (default `text`)
- `--summary <path>`
- `--no-color`
- `--dry-run`
- `--explain`

Canonical summary payload for every harness invocation:

- `schema_version`
- `command`
- `status` (`ok|fail`)
- `started_at`, `finished_at`, `duration_ms`
- `checks[]` (`name`, `status`, `details`)
- `failures[]` (`code`, `message`, `hint`, `retryable`)
- `artifacts[]`

Unified exit code mapping:

- `0` success
- `2` usage/args
- `3` missing dependency
- `4` contract validation failure
- `5` execution/test failure
- `6` file/io failure
- `7` internal/unexpected failure

Capability discovery:

- `agentcli loop capabilities --format json`
- returns commands, command-specific flags, artifact contracts, and schema version.

## Section 3: Migration Scope (No Compatibility) (Approved)

Aggressive migration rules:

1. Runtime behavior immediately moves to `tools/harness` as canonical source.
2. `cmd/agentcli/loop.go` retains CLI surface but delegates behavior entirely.
3. Legacy path is not preserved as fallback behavior.
4. Tests shift from text-heavy assertions to contract/golden assertions.
5. CI gates on schema stability and regression summary contract.

## Section 4: Behavior-Only Regression Lane (Approved)

Regression is a first-class command:

- `agentcli loop regression`
- `agentcli loop regression --write-baseline`
- optional `--profile` (default `quality`)
- optional `--baseline`

Behavior scope only (performance excluded for now):

- scenario step names + exit codes
- findings (`code`, `severity`, `source`)
- judge envelope (`pass`, `score`, `threshold`, `hard_failures`)
- committee strategy metadata (if present)

Policy:

- Any behavior drift fails command with contract failure exit semantics.
- Baseline updates are explicit (`--write-baseline`) and auditable in git.

Default baseline location:

- `testdata/regression/loop-<profile>.behavior-baseline.json`

## Non-Goals

- No performance regression framework in this phase.
- No compatibility wrapper for old ad-hoc output semantics.
- No partial migration where both engines co-own behavior.

## Risks and Mitigations

1. Risk: Big-bang cutover can break command assumptions.
   - Mitigation: contract golden tests for each command and strict exit-code matrix tests.
2. Risk: Summary schema drift across commands.
   - Mitigation: single contract package + schema fixtures in CI.
3. Risk: Human readability regresses when machine output is prioritized.
   - Mitigation: keep `text` renderer concise and deterministic with identical underlying contract data.

## Validation Criteria

Success criteria for completion:

1. All harness commands emit contract summaries without log scraping.
2. Failures are typed and machine-actionable with retry hints.
3. `capabilities` discovery works and is documented.
4. `task ci` fails on behavior drift and contract schema drift.
5. Existing regression lane remains behavior-only and deterministic.

## Decision Record

This design intentionally prioritizes a clean long-term architecture over compatibility.
Chosen because the requested direction is aggressive, immediate, and elegance-focused.
