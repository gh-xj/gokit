# Lean Commands

Use for daily checks with minimal cognitive load.

```bash
agentcli loop doctor --repo-root . --md
agentcli loop lean --repo-root .            # low-noise profile
agentcli loop quality --repo-root .         # strict skill package quality pass
agentcli loop autofix --repo-root . --threshold 9.0 --max-iterations 3
```
