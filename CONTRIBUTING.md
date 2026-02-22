# Contributing

Thanks for contributing to `agentcli-go`.

## Development Workflow

1. Fork and create a feature branch.
2. Make focused, deterministic changes.
3. Run local verification:

```bash
task ci
```

4. Open a PR with:
- problem statement
- approach summary
- verification evidence

## Contribution Rules

- Preserve deterministic scaffold/runtime behavior.
- If output contracts change, update:
  - `schemas/*.schema.json`
  - `testdata/contracts/*.ok.json`
  - `testdata/contracts/*.bad-*.json` when relevant
- Prefer small, reviewable PRs.

## Commit Style

Use concise conventional prefixes where practical:
- `feat:`
- `fix:`
- `docs:`
- `test:`
- `build:`
- `refactor:`

## Questions

Open a GitHub issue for design questions before large changes.
