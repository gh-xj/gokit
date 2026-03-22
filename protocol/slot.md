# Slot Protocol

Defines the slot/claim system for working directory isolation.

## Slot Identity

- Each slot is a directory copy whose identity is extracted from its basename minus the copy_prefix
- Slot names must match the pattern declared in strategy's slot.md
- Example: if copy_prefix is `myproject` and directory is `myproject-maxwell`, slot name is `maxwell`
- If copy_prefix is empty, the repo dirname is used at runtime

## Claim Protocol

- A case's "Claimed By" field tracks which slot owns it
- Only one slot may claim a case at a time
- Before working on a case, verify it is unclaimed or claimed by current slot
- If another slot owns the case, HALT (do not proceed)

## Slot Lifecycle

- Create: `casectl slot create <name>` → cp -r directory copy at `<parent>/<prefix>-<name>`
- List: `casectl slot list` → enumerate copy directories matching the prefix pattern
- Remove: `casectl slot remove <name>` → safety check + copy directory removal
- Sync: update slot copy from main branch (at explicit boundaries only)
- Branch naming: bare slot name (e.g. `maxwell`), no `slot/` prefix

Project-specific wrappers may automate these actions, but dispatcher-managed case flows must stay consistent with this contract.

## Strategy Configuration

Project's `.agentops/slot.md` declares:
- Naming pattern (regex)
- Path convention: `<parent>/<prefix>-<name>` (copy placed alongside the source repo)
- Sync policy (when to pull from main)
