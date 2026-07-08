# Polymetrics CLI Connector Parity

**Generated via:** official GSD Core Pi adapter command path
**Commands:** `scripts/gsd prompt onboard --fast --skip-phases`, `scripts/gsd prompt map-codebase --fast`, `scripts/gsd prompt new-project --from-existing --non-interactive`, `scripts/gsd prompt milestone-summary --planning-only`
**Upstream GSD Core:** `open-gsd/gsd-core@20297a8ff941378b8615a5d3e8629e52c10a0f9d`
**Runtime adapter:** `pi-project-local`

## What This Is

Polymetrics CLI (`pm`) is a local-first Go CLI for dependency-free ETL, reverse ETL, connector inspection, credential management, local warehouse queries, scheduling, and optional runtime-backed execution. This GSD Core brownfield project tracks connector parity across every documented connector surface, not only REST APIs.

Connector parity includes REST/JSON, GraphQL, XML/SOAP, CSV/NDJSON/report exports, binary transfers, file/object storage, SQL/CDC, queues/events/webhooks/audit logs, native protocols, direct-read commands, and product-safe writes.

## Core Value

Users and agents can trust `pm` as a connector-complete, safety-gated ETL and reverse ETL interface whose advertised connector capabilities match documented upstream product surfaces without duplicate or unsafe exposure.

## Current Repo Snapshot

Generated from the working tree during the GSD Pi refresh:

| Signal | Count |
|---|---:|
| Connector definition directories | 547 |
| Connector `api_surface.json` files | 547 |
| Stream definition files | 7159 |
| Write definition files | 5699 |
| Hook directories | 78 |
| Native connector directories | 37 |
| Go source files under `cmd/` + `internal/` | 491 |
| Repo-local YAML agent specs | 14 |
| Official GSD commands exposed through Pi adapter | 69 |

These counts are planning inputs only. Phase 1 remains the hard gate for regenerated, de-duplicated, authoritative inventory before connector fanout.

## Requirements

### Validated

- ✓ Single-binary Go CLI architecture exists with ETL, query, reverse ETL, scheduling, docs, and smoke verification.
- ✓ Connector Architecture v2 exists with embedded JSON definition bundles, declarative engine, hooks, native connectors, conformance, and code generation.
- ✓ Reverse ETL uses plan, preview, approval, and execute semantics; raw generic shell, generic HTTP write, and generic SQL write tools remain out of scope.
- ✓ Certification harness components exist under `internal/connectors/certify/` with source, write, replay, report, ledger, and sweeper stages.
- ✓ Official GSD Core command docs are pinned and surfaced through `scripts/gsd` plus Pi resources under `.pi/`.

### Active

- [ ] Reconcile the current connector inventory before any connector fanout.
- [ ] Preserve connector parity as the north-star milestone across GSD planning artifacts.
- [ ] Classify all documented connector surfaces across protocols without duplicating operations.
- [ ] Ensure every product-safe documented stream/report/event/log is available through `pm` CLI and ETL or a typed exclusion.
- [ ] Ensure every product-safe documented write/mutation is available through reverse ETL plan/preview/approval/run or a typed exclusion.
- [ ] Ensure direct-read and binary-transfer endpoints are supported where product-safe and human-gated where risky.
- [ ] Ensure conformance and certification gates enforce parity before rollout claims.
- [ ] Ensure agents and subagents use repo-local GSD commands via `.pi` or `scripts/gsd`, with manual fallback recorded only when the adapter is unavailable.
- [ ] Ensure CLI feature work updates runtime help, bare namespace help behavior, `docs/cli/**`, website docs, generated help/manual artifacts, and tests together.

### Out of Scope

- Runtime connector changes for issue #122 — this issue only reboots planning and agent guidance.
- New dependencies — dependency additions require explicit human approval.
- Credentialed live connector checks — no secrets are needed for planning rebootstrap.
- Destructive/admin execution — remains human-gated and outside automated planning edits.
- Generic shell, generic HTTP write, or generic SQL write tools — explicitly disallowed safety boundary.
- Phase regeneration in this refresh — user explicitly requested updating everything except phases.

## Context

This repository previously carried a custom/legacy `.planning/` tree with stale phase artifacts and connector counts. Issue #122 reboots active planning using official GSD Core workflow shape while preserving Polymetrics-specific connector parity overlays from `AGENTS.md`, `docs/migration/HANDOFF-CODEX.md`, `docs/migration/conventions.md`, `docs/architecture/connector-architecture-v2-design.md`, `docs/architecture/connector-certification-design.md`, and `docs/plans/universal-programming-loop-prd.md`.

The official GSD docs do not currently list Pi as an upstream runtime. This repository therefore provides a project-local Pi adapter that renders official command prompts through `scripts/gsd`, exposes `/gsd` plus generated `/gsd-*` aliases through `.pi/extensions/gsd/index.ts`, and loads default behavior through `.pi/skills/gsd-core/SKILL.md`.

## Constraints

- **Safety**: No secrets, credentialed checks, reverse ETL execution, destructive/admin actions, or production deploys during planning work.
- **Source scope**: Planning-only issue; `cmd/` and `internal/` must not change.
- **Human gates**: New dependencies, auth scope changes, destructive/admin operations, quality-gate reductions, and merges to `main` require human approval.
- **Architecture**: Connector runtime remains declarative-first with Tier 2 hooks and Tier 3 natives only when justified by `docs/migration/conventions.md`.
- **Surface coverage**: Connector parity covers REST, GraphQL, XML/SOAP, report exports, binary transfers, file/object storage, SQL/CDC, queues/events/webhooks, native protocols, direct-read, and writes.
- **De-duplication**: One upstream operation maps to exactly one primary classification; aliases and duplicate docs references are cross-links, not duplicated work.
- **Agent runtime**: Agents and subagents should prefer `/gsd <command>` or `scripts/gsd prompt <command>` from the repo-local Pi adapter.
- **CLI docs parity**: CLI-visible changes must follow `.agents/agentic-delivery/references/cli-help-docs-website-parity.md` so runtime help, manual docs, website docs, and generated artifacts stay aligned.
- **Review**: Issue-to-PR delivery follows `.agents/agentic-delivery/contracts/issue-agent-contract.md`; PR targets `main` and uses `Closes #122`.

## Key Decisions

| Decision | Rationale | Outcome |
|---|---|---|
| Rebootstrap active `.planning/` from official GSD Core workflow shape | The old tree was custom/legacy and should not pollute future onboarding | In progress |
| Pin official `open-gsd/gsd-core@next` docs for Pi adapter | Pi is not listed as an upstream runtime, so the repo needs explicit provenance for its adapter | Good |
| Expose GSD through `.pi` and `scripts/gsd` | Agents need runtime-neutral command access and Pi slash-command aliases | Good |
| Require CLI help/manual/website parity | CLI features should be discoverable via `pm help`, bare namespace commands like `pm connectors`, `docs/cli/**`, website docs, and generated help artifacts | Good |
| Archive old `.planning/` outside active planning context | Preserves auditability without tracking stale artifacts | Good |
| Inventory reconciliation precedes connector fanout | Current bundle/API/certification/surface counts must be trusted before parallel work | Pending |
| Treat connector parity as multi-surface, not REST-only | GraphQL, XML/SOAP, CSV/NDJSON, binary, file, database, CDC, queue, webhook, and direct-read surfaces exist in the repo | Pending |
| Keep active `.planning/` tracked | Issue explicitly requires upstream-generated planning to remain versioned | Good |

---
*Last updated: 2026-07-08 via repo-local official GSD Core Pi adapter; phases intentionally not regenerated in this refresh.*
