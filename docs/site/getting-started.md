# Getting Started

## Install

Homebrew (recommended):

```bash
brew tap gh-xj/tap
brew install agentcli
```

Or with Go:

```bash
go install github.com/gh-xj/agentcli-go/cmd/agentcli@v0.2.0
```

## Install Verification

```bash
which agentcli
agentcli --version
agentcli --help
```

## Scaffold

```bash
agentcli new --module example.com/mycli mycli
agentcli add command --dir ./mycli --preset file-sync sync-data
agentcli doctor --dir ./mycli --json
```

Available presets:

```bash
agentcli add command --list-presets
```

## Verify

```bash
cd mycli
task verify
```

## First 5 Minutes Script

```bash
set -e
brew tap gh-xj/tap
brew install agentcli

mkdir -p /tmp/agentcli-demo && cd /tmp/agentcli-demo
agentcli new --module example.com/demo demo
agentcli add command --dir ./demo --preset file-sync sync-data
agentcli doctor --dir ./demo --json
cd demo && task verify
```

## Onboarding Benchmark

Current baseline from partner onboarding reports:

- First scaffold success: ~1 minute
- First `task verify` pass: ~1 minute
- `doctor` iterations before green: 1
