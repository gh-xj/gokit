# Hooks Protocol

Defines lifecycle hook points that strategy can bind to.

## Available Hooks

| Hook | When Fired | Use Cases |
|------|-----------|-----------|
| on-case-open | After case record created | Create external tracker issue, notify team |
| on-case-transition | Status changes | Sync to external tracker, trigger dependent workflows |
| on-worker-complete | A worker finishes | Trigger dependent workers, update progress |
| on-reconcile-done | After all workers reconciled | Append evolution backlog, run audits |
| on-case-close | Before final commit on resolved/closed | Trigger reflection, cleanup worktrees, archive |

## Hook Definition

In strategy's hooks.md, each hook maps to one or more actions:
- Skill invocation (invoke a named skill)
- CLI command (run a casectl or project command)
- Shell command (run arbitrary command)

## Execution

- Hooks are non-blocking by default
- Hook failures are logged but do not block the dispatch cycle
- Strategy can mark hooks as blocking via `blocking: true`
