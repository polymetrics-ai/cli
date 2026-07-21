# Phase Agents

Phase: `471-pi-agent-session-shepherd`

The root agent owns issue state, planning artifacts, TDD ordering, implementation integration,
verification, GitHub delivery, and the final human-gate handoff. Read-only scouts may inspect
independent concerns; no mutating worker shares this checkout.

| Role | Mode | Write scope |
| --- | --- | --- |
| coordinator/implementer | local critical path in isolated issue worktree | issue phase, `.pi/extensions/shepherd/**`, `.pi/settings.json`, `.pi/README.md` |
| Pi infrastructure scout | read-only subagent | none |
| Pi SDK scout | read-only subagent | none |
| GitHub topology scout | read-only subagent | none |
| final reviewer | read-only subagent after exact-head verification | none |
