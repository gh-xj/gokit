# CI Failure Example

## Event

```json
{
  "schema_version": "dogfood-event.v1",
  "event_id": "evt-2026-03-01-ci-001",
  "event_type": "ci_failure",
  "signal_source": "github_actions",
  "timestamp": "2026-03-01T11:45:00Z",
  "repo_guess": "gh-xj/agentcli-go",
  "error_summary": "task docs:check failed on docs drift",
  "evidence_paths": [
    ".github/workflows/ci.yml",
    "skills/verification-loop/SKILL.md"
  ]
}
```

## Commands

```bash
task dogfood:dry-run EVENT=.docs/dogfood/event.ci.json
task dogfood:publish EVENT=.docs/dogfood/event.ci.json
```

## Expected result

- Repeated CI failures with the same fingerprint dedupe into one open issue thread.
- Ledger status remains replayable for future retries if publish fails.
