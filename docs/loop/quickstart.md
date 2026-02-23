# Loop Quickstart

Use these three commands for daily verification with low cognitive load.

## 1) Run

```bash
agentcli loop run --repo-root . --threshold 9.0 --max-iterations 1
```

## 2) Judge Gate

```bash
agentcli loop judge --repo-root . --threshold 9.0 --max-iterations 1
```

## 3) Autofix

```bash
agentcli loop autofix --repo-root . --threshold 9.0 --max-iterations 3
```

## Review Output

Primary reviewer file:

- `.docs/onboarding-loop/review/latest.md`

Or print it directly:

```bash
agentcli loop review --repo-root .
```

JSON output for automation:

```bash
agentcli loop review --repo-root . --json
```

Primary machine-readable summary:

- `.docs/onboarding-loop/latest-summary.json`

Run retention:

- keeps most recent 20 runs under `.docs/onboarding-loop/runs/`

## Optional Lab Mode

Use advanced workflows only when needed:

- `agentcli loop lab compare ...`
- `agentcli loop lab replay ...`
- `agentcli loop lab run --verbose-artifacts ...`

Reviewer checklist:

- `skills/verification-loop/CHECKLIST.md`

## Troubleshooting in 60s

1. `loop doctor` fails lean-ready:
   - run `agentcli loop doctor --repo-root . --md`
   - fix onboarding install prompt issues reported in findings.
2. `loop review` says file missing:
   - run `agentcli loop judge --repo-root . --max-iterations 1` once to generate review output.
3. replay/compare needs more artifacts:
   - run `agentcli loop lab run --repo-root . --max-iterations 1 --verbose-artifacts`.
