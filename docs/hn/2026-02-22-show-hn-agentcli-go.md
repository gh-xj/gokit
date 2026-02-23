# Show HN: agentcli-go

## Title Options

1. Show HN: agentcli-go — deterministic Go CLIs for AI agents
2. Show HN: agentcli-go — scaffold Go CLIs with contract-based verification
3. Show HN: agentcli-go — from prompt to production CLI with CI contracts

## Short Post (Recommended)

I built `agentcli-go`, a framework for generating and maintaining deterministic Go CLIs that AI agents can safely evolve.

Core ideas:
- strict scaffolded layout
- machine-readable health checks (`agentcli doctor --json`)
- schema-validated smoke outputs
- CI gates that enforce output contracts (including negative regression checks)

Why:
I wanted agent-generated CLI workflows with Go-level determinism and compile-time safety, not script drift.

What I want feedback on:
1. Is onboarding clear enough from zero to first working command?
2. Is strictness at the right level, or too opinionated?
3. Which extension points matter most for real-world adoption?

Quick start:
```bash
go install github.com/gh-xj/agentcli-go/cmd/agentcli@v0.2.0
agentcli new --module example.com/mycli mycli
agentcli add command --dir ./mycli --description "sync local files" sync-data
agentcli doctor --dir ./mycli --json
cd mycli && task verify
```

Current onboarding baseline (internal partner runs):
- first scaffold success: ~1 minute
- first `task verify` pass: ~1 minute
- median `doctor` iterations before green: 1

Punch line:
Deterministic Go CLIs that AI agents can safely evolve.

## Longer Post

`agentcli-go` is a Go CLI framework focused on a specific problem:
How do you let AI agents create and modify CLIs without causing long-term drift?

I built it around explicit contracts:
- scaffolded project shape is predictable
- command outputs support machine mode (`--json`)
- health checks are parseable (`doctor --json`)
- smoke artifacts are schema-validated
- CI includes negative fixtures so bad outputs are expected to fail

This makes generated CLIs easier to trust in automation pipelines.

I am especially looking for feedback on:
1. first-time onboarding friction,
2. strictness vs flexibility,
3. gaps in extension model for OSS usage.

If useful, I can share a before/after migration example from skill-driven script repos.

## Reference Link Usage

If you reference external interviews or ideas (for example Lex Fridman transcripts), keep it short:
- paraphrase the idea in your own words,
- include the source link directly,
- avoid long quotes.
