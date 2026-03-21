# Worker Protocol

Defines the registration contract for case workers.

## Worker Registration

Workers may be declared in either:

- legacy `.agentops/workers/<name>/SKILL.md`
- project-local `.claude/skills/<name>/SKILL.md`

The worker skill must declare frontmatter:

```yaml
---
worker-type: review | verify | challenge | reflect | triage | custom
sidecar-path: <relative-path-from-case-dir>
blocking: true | false
requires: [<other-worker-names>]
capabilities: [read-only, can-edit, can-run-commands]
---
```

## Fields

- **worker-type**: Classification of what this worker does
- **sidecar-path**: Where the worker writes its output (relative to case directory)
- **blocking**: Whether case closure depends on this worker completing
- **requires**: Workers that must complete before this one starts (sequencing)
- **capabilities**: What the worker is allowed to do

## Constraints

- No circular dependencies in `requires` graph
- Each worker must have a unique sidecar-path
- Workers must not write to case.md directly
- Workers must not modify other workers' sidecars
- `.agentops/worker-registry.md` may summarize workers, but worker skill frontmatter is the source of truth
