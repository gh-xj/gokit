# agentcli-go

Deterministic Go CLI framework for human + AI-agent teams.

Scaffold fast, verify by contract, ship with confidence.

## Why agentcli-go

Most CLI projects start fast and drift later.
`agentcli-go` is built to prevent that drift from day one.

- Deterministic scaffolding for every new CLI
- Stable machine-readable checks (`doctor --json`)
- Schema-validated smoke artifacts
- CI gates that catch regressions early (including negative checks)
- Go compile-time safety instead of script-time surprises

## 60-Second Quickstart

Generate a new CLI project:

```bash
go run ./cmd/agentcli new --module example.com/mycli mycli
```

Add a command:

```bash
go run ./cmd/agentcli add command --dir ./mycli sync-data
```

Run health checks:

```bash
go run ./cmd/agentcli doctor --dir ./mycli --json
```

Run project verification:

```bash
cd mycli
task verify
```

## AI-Agent Onboarding

If you are using coding agents, this framework gives them explicit contracts to follow.

Recommended agent workflow:

1. `agentcli new` to initialize a contract-compliant baseline.
2. `agentcli add command` for feature growth.
3. `agentcli doctor --json` before and after changes.
4. `task ci` as the canonical pass/fail gate.

Why this works well for agents:

- predictable file layout
- deterministic command output paths
- schema-backed JSON outputs
- low ambiguity in verification steps

## What You Get

### 1) Scaffold CLI

- `agentcli new`
- `agentcli add command`
- `agentcli doctor --json`

### 2) Runtime Modules

- `cobrax`: standardized Cobra root flags + deterministic exit-code handling
- `configx`: deterministic config merge order (`Defaults < File < Env < Flags`)

### 3) Core Helpers

- logging: `InitLogger()`
- args: `ParseArgs`, `RequireArg`, `GetArg`, `HasFlag`
- exec: `RunCommand`, `RunOsascript`, `Which`, `CheckDependency`
- fs: `FileExists`, `EnsureDir`, `GetBaseName`
- runtime contracts: `NewAppContext`, `RunLifecycle`, `NewCLIError`, `ResolveExitCode`

## Generated Project Contract

Each generated project includes:

- fixed scaffold layout
- deterministic smoke artifact at `test/smoke/version.output.json`
- schema file at `test/smoke/version.schema.json`
- canonical task gates (`fmt`, `lint`, `test`, `build`, `smoke`, `ci`, `verify`)

Example structure:

```text
mycli/
  cmd/
    root.go
    <command>.go
  internal/
    app/
    config/
    io/
    tools/smokecheck/
  pkg/version/
  test/e2e/
  test/smoke/
  Taskfile.yml
  main.go
```

## Verification Model

This repository enforces JSON output contracts in CI.

- schemas: `schemas/doctor-report.schema.json`, `schemas/scaffold-version-output.schema.json`
- positive fixtures: `testdata/contracts/*.ok.json`
- negative fixtures: `testdata/contracts/*.bad-*.json`
- gates: `task schema:check`, `task schema:negative` (both part of `task ci`)

## Install as a Library

```bash
go get github.com/gh-xj/agentcli-go@latest
```

Example usage:

```go
package main

import (
    agentcli "github.com/gh-xj/agentcli-go"
    "github.com/rs/zerolog/log"
)

func main() {
    agentcli.InitLogger()
    args := agentcli.ParseArgs([]string{"--src", "/tmp/in"})
    src := agentcli.RequireArg(args, "src", "source directory")
    log.Info().Str("src", src).Msg("ready")
}
```

## Design Principles

- Determinism over convenience
- Contracts over conventions
- Explicit verification over implicit correctness
- Agent-friendly interfaces without sacrificing human ergonomics

## Contributing

Issues and PRs are welcome.

When contributing:

1. Keep scaffold/runtime contracts deterministic.
2. Add or update schema fixtures when output contracts change.
3. Run `task ci` before opening PRs.
