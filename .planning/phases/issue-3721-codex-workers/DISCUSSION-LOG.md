# DISCUSSION LOG — issue #3721 Codex project-local workers

Command path: `scripts/gsd prompt discuss-phase issue-3721-codex-workers --auto`.

No open product choices remained after the issue, parent contract, and task constraints were read.
The auto mode records the locked implementation decisions in `CONTEXT.md` rather than reopening
them. The official Codex documentation was re-verified before planning.

| Area | Decision | Evidence |
|---|---|---|
| Format and discovery | Project-scoped standalone TOML under `.codex/agents/` | Codex custom-agents documentation |
| Delegation isolation | Set `agents.enabled = false` in every generated role | Codex global subagent-settings documentation |
| Global/project collisions | Do not rely on undocumented same-filename precedence | Official docs document built-in-name override only |
| Trust | Trusted project is a prerequisite; untrusted skips project `.codex/` layers | Codex config basics |
| Legacy project roles | Retain for Wave 6, per active task scope fence | Task brief |
