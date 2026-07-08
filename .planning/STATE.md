# Project State

**Project:** Polymetrics CLI Connector Parity
**Last activity:** 2026-07-08 — GSD and agent guidance updated so agents/subagents load required Go/design skills and CLI feature work keeps runtime help, bare namespace help behavior, `docs/cli/**`, website docs, generated help/manual artifacts, and tests in parity.

## Current State

- Issue #122 is active on branch `chore/122-upstream-gsd-rebootstrap` with PR #123 open.
- Active `.planning/` replaces a legacy/custom tree; previous active planning is archived outside the current tree.
- Official `open-gsd/gsd-core@next` docs are pinned in `.gsd/upstream.lock.json` and adapted for Pi through `scripts/gsd` plus `.pi/` resources.
- `.gsd/commands.json` exposes 69 official GSD commands generated from official `docs/COMMANDS.md`.
- `.pi/extensions/gsd/index.ts` exposes `/gsd` plus generated `/gsd-*` aliases after project trust/reload.
- `.pi/skills/gsd-core/SKILL.md` provides default GSD behavior for Pi.
- `.agents/**` guidance routes GSD work through the Pi adapter or `scripts/gsd prompt`.
- `.agents/agentic-delivery/references/required-skills-routing.md` defines required Go/design skill routing for agents and subagents.
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md` defines the required parity gate for CLI-visible changes.
- Connector parity includes REST, GraphQL, XML/SOAP, CSV/NDJSON, binary, file/object, SQL/CDC, queues/events/webhooks, native protocols, direct-read, and writes.
- Phase 1 inventory reconciliation is a hard gate before connector fanout.
- `.planning/phases/**` was intentionally not regenerated in this refresh per user request.

## Current Quick Inventory Inputs

| Signal | Count |
|---|---:|
| Connector definition directories | 547 |
| Connector `api_surface.json` files | 547 |
| Stream definition files | 7159 |
| Write definition files | 5699 |
| Hook directories | 78 |
| Native connector directories | 37 |
| Go files under `cmd/` + `internal/` | 491 |
| YAML agent specs | 14 |
| GSD commands in `.gsd/commands.json` | 69 |

These are quick-map inputs only; authoritative counts require Phase 1 inventory reconciliation.

## Active Decisions

- Use official GSD Core docs as source of truth for workflow commands.
- Treat Pi as a project-local adapter target, not an upstream-supported runtime.
- Prefer `/gsd <command>` or generated `/gsd-*` aliases in Pi.
- Prefer `scripts/gsd prompt <command>` for deterministic traces and non-interactive automation.
- Keep manual-GSD fallback only for adapter-unavailable cases and record it explicitly.
- Keep `cmd/`, `internal/`, and `.planning/phases/**` unchanged for the current non-phase refresh.
- For future CLI feature work, require parity across `pm help <topic>`, bare namespace invocations such as `pm connectors`, `pm <command> --help`, `docs/cli/**`, `website/**`, generated help/manual artifacts, and tests.
- For future Go work, require `golang-how-to` plus task-specific Go skills. For website/docs UI work, require applicable design skills such as `frontend-design`, `web-design-guidelines`, and `vercel-react-best-practices`.

## Blockers / Human Gates

- Do not use live connector credentials for issue #122.
- Do not add dependencies without human approval.
- Do not execute reverse ETL during planning.
- Do not run destructive/admin/elevated external actions.
- Do not merge PR #123 to `main` without human approval.

## Next Expected Work

1. Run GSD/Pi, YAML/JSON, diff, and scope verification.
2. Commit and push the CLI help/docs/website parity guidance to PR #123.
3. Let CI and CodeRabbit automatic review run; do not post redundant manual review commands unless documented fallback conditions apply.

---
*State refreshed: 2026-07-08 via repo-local official GSD Core Pi adapter.*
