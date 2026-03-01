# Local Runtime Error Example

## Event

```json
{
  "schema_version": "dogfood-event.v1",
  "event_id": "evt-2026-03-01-runtime-001",
  "event_type": "runtime_error",
  "signal_source": "local",
  "timestamp": "2026-03-01T11:20:00Z",
  "repo_guess": "gh-xj/agentcli-go",
  "error_summary": "panic: nil pointer dereference",
  "evidence_paths": [
    "internal/tools/dogfoodfeedback/main.go:120"
  ]
}
```

Fingerprint is intentionally omitted from input; the tool derives it at runtime.

## Commands

```bash
task dogfood:dry-run EVENT=.docs/dogfood/event.runtime.json
task dogfood:publish EVENT=.docs/dogfood/event.runtime.json
```

## Expected result

- Dry-run prints a decision, fingerprint, repo, and confidence.
- Publish creates an issue (or comments on an existing one) and appends a ledger record.
