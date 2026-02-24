# Release Playbook

Use this checklist for every tag publish.

## Mandatory gate before tag publish

```bash
task release:gate
```

This includes:

- full CI (`task ci`)
- core-only compile (`go build -tags agentcli_core ./cmd/agentcli`)
- release gate command check (`go run ./cmd/agentcli --help`)

## Publish

```bash
task release:publish TAG=vX.Y.Z NOTES=docs/releases/vX.Y.Z.md
```

## Proxy propagation fallback

If module proxy propagation is delayed or unavailable, retry release/install validation with direct proxy mode:

```bash
GOPROXY=direct task release:gate
GOPROXY=direct go install github.com/gh-xj/agentcli-go/cmd/agentcli@vX.Y.Z
```

## Monorepo recommendation

For monorepo consumers, keep `--in-existing-module` as the default recommendation:

```bash
agentcli new --dir ./tools --in-existing-module my-tool
```
