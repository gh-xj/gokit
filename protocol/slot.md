# Slot Protocol

Defines the slot/claim system for working directory isolation.

## Slot Identity

- Each slot has a `.slot` marker file at its worktree root
- Content: exactly the slot name (no newline)
- Slot names must match the pattern declared in strategy's slot.md

## Claim Protocol

- A case's "Claimed By" field tracks which slot owns it
- Only one slot may claim a case at a time
- Before working on a case, verify it is unclaimed or claimed by current slot
- If another slot owns the case, HALT (do not proceed)

## Slot Lifecycle

- Create: `casectl slot create <name>` → git worktree + .slot marker
- List: `casectl slot list` → enumerate worktrees with .slot markers
- Remove: `casectl slot remove <name>` → safety check + worktree removal
- Sync: update slot worktree from main branch (at explicit boundaries only)

Project-specific wrappers may automate these actions, but dispatcher-managed case flows must stay consistent with this contract.

## Strategy Configuration

Project's `.agentops/slot.md` declares:
- Naming pattern (regex)
- Path convention (worktree location)
- Sync policy (when to pull from main)
