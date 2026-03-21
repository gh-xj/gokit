# Case Record Protocol

Defines the minimum interface for a case record (case.md).

## Required Sections

### Metadata (required)

Every case.md must have a Metadata section with at minimum:
- **Type**: one of `intake`, `pr`, `quality`, `mixed`, or a custom type
- **Status**: one of `open`, `in_progress`, `resolved`, `closed_no_action` (extensible)
- **Created**: YYYY-MM-DD date

### Status Enum

Required values: `open`, `in_progress`, `resolved`, `blocked`, `closed_no_action`

Projects may add custom statuses (e.g., `needs_user_input`, `needs_rebase`).

### Status Groups

Group aliases for CLI filtering (`casectl case list --status <group>`):

| Group | Expands To |
|-------|------------|
| `active` | `open`, `in_progress`, `blocked` |
| `completed` | `resolved`, `closed_no_action` |

### Directory Organization

Cases are stored in `{group}/{slot}/CASE-*` subdirectories:

| Directory | Contains |
|-----------|----------|
| `cases/active/<slot>/` | open, in_progress, blocked |
| `cases/completed/<slot>/` | resolved, closed_no_action |

Slots are discovered dynamically by scanning subdirectories — no hardcoded list.

When status changes cross storage groups, a dispatcher or compatible case tool may move the case directory while preserving the slot: `active/<slot>/CASE-X` → `completed/<slot>/CASE-X`. The `- Status:` field in case.md remains the source of truth.

## Extension Points

Strategy's schema.md may add any additional sections and metadata fields. Common extensions:
- Claimed By / Claimed At (for slot ownership)
- Risk Classification
- Active Workflows
- Findings (with links to worker sidecars)
- Artifacts
- Next Action / Open Questions / Close Criteria
- Linear-Ref / external tracker references

## Ownership

- Dispatcher owns case.md writes
- Workers write to sidecars, not directly to case.md
- Workers may recommend status changes; dispatcher decides
