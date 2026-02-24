# CI Commands

Use for strict quality gates in CI or scheduled workflows.

```bash
task ci
agentcli loop doctor --repo-root .
agentcli loop quality --repo-root .
agentcli loop regression --repo-root .
agentcli loop judge --repo-root . --threshold 9.0 --max-iterations 1
go run ./internal/tools/loopbench --mode run --repo-root . --output .docs/onboarding-loop/benchmarks/latest.json
go run ./internal/tools/loopbench --mode check --output .docs/onboarding-loop/benchmarks/latest.json --baseline testdata/benchmarks/loop-benchmark-baseline.json
```
