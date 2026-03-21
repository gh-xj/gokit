# Lifecycle Protocol

Defines the mandatory phases of a dispatch cycle.

## Phases

| Phase | Framework Owns | Strategy Provides |
|-------|---------------|-------------------|
| detect-slot | Read .slot, resolve case root | — |
| find-or-create | Locate/create case directory | schema.md (template) |
| classify | Determine case type | routing.md (classification cues) |
| assess-risk | Compute risk level | risk.md (triggers) |
| select-workers | Map (type, risk) → workers | budget.md + routing.md |
| execute-workers | Launch workers, collect sidecars | worker skills from `.agentops/workers/` or `.claude/skills/` |
| reconcile | Merge worker outputs into case | routing.md (reconciliation rules) |
| fire-hooks | Execute lifecycle hooks | hooks.md |
| commit | One dispatcher-owned commit | — |

## Ordering

Phases execute in order. A phase may be skipped if its strategy file is absent (using defaults).

## Error Handling

If a phase fails:
- Log the error in the case record
- Set status to `blocked` if the failure is unrecoverable
- Do not skip subsequent phases silently
