---
name: dogfood-feedback-loop
description: Use when capturing dogfood failures and routing deduplicated feedback to GitHub issues with local replay support.
version: 1.0
---

# dogfood-feedback-loop Skill

## In scope

- Capture CI failure, runtime error, and docs drift events in a canonical JSON format.
- Route feedback to the best repo target using manual override, git remote, and cwd inference.
- Dedupe repeated failures by fingerprint and persist local ledger history.
- Publish feedback to GitHub as a new issue or comment on an existing issue.

## Use this when

- You need repeatable dogfood failure intake instead of ad-hoc notes.
- You need issue dedupe across repeated failures.
- You need a dry-run safety check before publishing.
- You need to replay queued events after auth or network recovery.

## Feedback workflow

1. Verify GitHub CLI authentication:
   - `gh auth status`
   - If needed: `gh auth login`
2. Prepare an event JSON file (schema: `schemas/dogfood-event.schema.json`).
3. Dry-run the decision:
   - `task dogfood:dry-run EVENT=.docs/dogfood/event.json`
4. Publish feedback:
   - `task dogfood:publish EVENT=.docs/dogfood/event.json`
5. Check local state:
   - Ledger: `.docs/dogfood/ledger.json`
   - Idempotency marker: `.docs/dogfood/ledger.idempotency.json`

## Replay

- Replay by re-running the tool for the same event after fixing local conditions (auth/network/repo target).
- If an open issue already exists for the fingerprint, replay appends a comment instead of creating a duplicate issue.
- Minimal replay command:
  - `go run ./internal/tools/dogfoodfeedback --event .docs/dogfood/event.json --ledger .docs/dogfood/ledger.json --repo gh-xj/agentcli-go`
- Use `--dry-run` first if current ledger state is unknown.

## Examples

- [Local runtime error](./examples/local-runtime-error.md)
- [CI failure](./examples/ci-failure.md)

## Out of scope

- This skill does not define triage ownership or issue label taxonomy.
- This skill does not replace repository verification gates (`task ci`, `task verify`).
