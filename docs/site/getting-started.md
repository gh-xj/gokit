# Getting Started

## Install

```bash
go install github.com/gh-xj/agentcli-go/cmd/agentcli@v0.2.0
```

## Scaffold

```bash
agentcli new --module example.com/mycli mycli
agentcli add command --dir ./mycli sync-data
agentcli doctor --dir ./mycli --json
```

## Verify

```bash
cd mycli
task verify
```
