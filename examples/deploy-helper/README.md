# deploy-helper example

Use-case: deployment helper with explicit preflight and postflight checks.

Typical commands:
- `preflight`
- `deploy`
- `rollback`

Contract notes:
- no implicit side effects before preflight success
- command output suitable for CI and agent orchestration
