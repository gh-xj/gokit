# AI Agent Playbook

Use this flow for deterministic agent-driven CLI development:

1. Bootstrap with `agentcli new`.
2. Add commands with `agentcli add command`.
3. Validate before and after edits: `agentcli doctor --json`.
4. Enforce quality gate: `task ci`.

If command output contracts change, update schema fixtures and tests in the same PR.
