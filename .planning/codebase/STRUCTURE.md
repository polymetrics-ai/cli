# Structure

**Generated via:** `scripts/gsd prompt map-codebase --fast` through the official GSD Core Pi adapter.

## Repository Areas

| Path | Purpose |
|---|---|
| `cmd/` | CLI entrypoints and command surfaces. Not edited for issue #122. |
| `internal/` | Product implementation, connector runtime, ETL/reverse ETL, state, warehouse, certification. Not edited for issue #122. |
| `internal/connectors/defs/` | JSON connector definition bundles. Current quick count: 547 directories. |
| `internal/connectors/hooks/` | Hook implementations for connector behavior not captured declaratively. Current quick count: 78 directories. |
| `internal/connectors/native/` | Native protocol implementations. Current quick count: 37 directories. |
| `docs/architecture/` | Architecture and certification design docs. |
| `docs/migration/` | Connector architecture v2 handoff, conventions, status, and wave guidance. |
| `.agents/` | Agent-neutral contracts, reusable YAML specs, task/skill matrix, and connector migration agents. |
| `.gsd/` | Official GSD source lock, docs snapshot, command registry, and repo-specific prompt sources. |
| `.pi/` | Pi settings, extension, prompt, and skill resources for official GSD command execution. |
| `.planning/` | Active GSD planning tree for issue #122. |
| `scripts/` | Repo-local scripts including `scripts/gsd`. |

## Important Generated/Planning Boundaries

- `.planning/phases/**` is intentionally unchanged in this refresh.
- `cmd/**` and `internal/**` are out of scope for issue #122 planning-only changes.
- `.gsd/official-docs/**` is an official docs snapshot pinned by `.gsd/upstream.lock.json`.
- `.gsd/commands.json` is generated from official `docs/COMMANDS.md`.
- `.pi/extensions/gsd/index.ts` registers `/gsd` and generated `/gsd-*` aliases.

## Agent Structure

- `.agents/agentic-delivery/agents/**/*.agent.yaml`: reusable delivery/orchestration/review/security/architecture agents.
- `.agents/connector-migration/agents/**/*.agent.yaml`: connector migration implementation/review agents.
- All agents should route GSD workflows through `.pi` when interactive and `scripts/gsd prompt` when non-interactive.

---
*Structure refreshed: 2026-07-08 via repo-local official GSD Core Pi adapter.*
