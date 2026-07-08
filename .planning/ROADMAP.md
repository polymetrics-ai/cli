# Roadmap: Polymetrics CLI Connector Parity

**Generated via:** official GSD Core Pi adapter command path
**Commands:** `scripts/gsd prompt onboard --fast --skip-phases`, `scripts/gsd prompt new-project --from-existing --non-interactive`, `scripts/gsd prompt milestone-summary --planning-only`
**Upstream GSD Core:** `open-gsd/gsd-core@20297a8ff941378b8615a5d3e8629e52c10a0f9d`
**Phase regeneration:** skipped by request

## Overview

This roadmap replaces the legacy custom `.planning/` tree with an official GSD Core brownfield plan for connector parity. The repo-local Pi adapter is the command path for future GSD work: agents use `/gsd <command>` or generated `/gsd-*` aliases in Pi, and non-interactive automation uses `scripts/gsd prompt <command>`.

The first delivery gate remains inventory reconciliation. No connector fanout starts until the repository and upstream documentation are reconciled across all connector technologies, canonical operation de-duplication is applied, and human-gated risk classification is reviewed.

## North-Star Milestone

Deliver connector parity that is:

1. **Surface-complete** — covers REST, GraphQL, XML/SOAP, CSV/NDJSON/report export, binary, file/object, SQL/CDC, queue/event/webhook/audit-log, native protocol, direct-read, and reverse ETL write surfaces.
2. **Safety-gated** — keeps secrets, auth scopes, destructive/admin operations, live credential checks, reverse ETL execution, new dependencies, quality-gate reductions, and `main` merges under human control.
3. **De-duplicated** — assigns each documented upstream operation exactly one primary classification.
4. **Conformance-backed** — derives parity claims from generated inventory, fixture/replay gates, conformance checks, and certification status.
5. **Agent-runnable** — uses official GSD Core commands through `.pi` and `scripts/gsd`, not runtime-specific copied command files.
6. **Skill-routed** — agents and subagents load required Go/design skills such as `golang-how-to`, `golang-cli`, `golang-testing`, `golang-security`, `golang-documentation`, `frontend-design`, `web-design-guidelines`, and `vercel-react-best-practices` as applicable.
7. **Help/docs/website-complete** — every CLI-visible feature keeps runtime help, bare namespace command behavior, `docs/cli/**`, website docs, generated help/manual artifacts, and tests in parity.

## Workstreams

### 0. GSD Runtime and Agent Enablement

**Goal:** Make the official GSD command surface the default for humans, Pi sessions, and reusable agents.

**Status:** In progress on issue #122.

**Required outcomes:**

- `.gsd/upstream.lock.json` pins official GSD source.
- `.gsd/commands.json` lists official commands generated from official docs.
- `.pi/extensions/gsd/index.ts` exposes `/gsd` plus `/gsd-*` aliases.
- `.pi/skills/gsd-core/SKILL.md` sets default planning/implementation behavior.
- `.agents/**` instructions route agents/subagents through the Pi adapter or `scripts/gsd`.
- `.agents/agentic-delivery/references/required-skills-routing.md` defines required Go/design skill routing for agents and subagents.
- `.agents/agentic-delivery/references/runtime-rlm-website-integration.md` preserves runtime/RLM/Pi-agent/website integration knowledge for Podman, PostgreSQL, DragonflyDB/Redis-compatible coordination, Temporal, RLM agent mode, and website docs.
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md` defines the CLI help/manual/website parity gate.
- Manual-GSD fallback is only used when the adapter is unavailable and must be recorded.
- Runtime-backed checks remain optional and gated; dependency-free CLI paths must not require PostgreSQL, DragonflyDB, Temporal, or Podman.

### 1. Inventory and Surface Reconciliation

**Goal:** Produce a current, generated, de-duplicated connector parity baseline before fanout.

**Depends on:** Workstream 0 for command/runtime consistency.

**Success criteria:**

- Active `.planning/` is official GSD Core shaped and the pre-rebootstrap archive is recorded outside active planning.
- Inventory reconciles bundles, hooks, natives, docs, API/surface manifests, streams, writes, binary surfaces, direct-read surfaces, native protocol surfaces, blockers, quarantine, conformance, and certification.
- Inventory classifies REST, GraphQL, XML/SOAP, CSV/NDJSON, binary, file/object, SQL/CDC, queue/event/webhook/audit-log, native, direct-read, and mutation surfaces.
- Canonical operation identity avoids duplicate work from multiple docs pages/specs.
- Connector fanout is blocked until inventory outputs are reviewed.

### 2. Durable Read and ETL Parity

**Goal:** Bring product-safe documented durable record collections to CLI and ETL parity.

**Depends on:** Workstream 1.

**Success criteria:**

- Every product-safe documented stream/report/event log/feed/table/queue message/CDC event/durable record collection is covered by a connector read surface or typed exclusion.
- `pm connectors inspect`, catalog, read, and ETL surfaces agree on stream names, schemas, sync modes, cursor behavior, and limitations.
- Incremental, pagination, cursor, schema, and projection behavior is verified through conformance fixtures or typed blockers.
- CLI-visible read surfaces update runtime help, bare namespace summaries, `docs/cli/**`, website docs, generated help/manual artifacts, and tests.

### 3. Direct-Read, Binary, and Native Surface Parity

**Goal:** Cover product-safe non-stream and non-REST read surfaces without forcing them into the wrong abstraction.

**Depends on:** Workstream 1.

**Success criteria:**

- Direct-read operations are classified separately from durable ETL streams.
- Binary operations have product-safe transfer surfaces or typed exclusions.
- GraphQL, XML/SOAP, CSV/NDJSON/report export, file/object, SQL/CDC, queue/event/webhook/audit-log, and native protocol reads use an appropriate declarative, hook, native, direct-read, binary, or exclusion path.
- CLI help/manual/website docs explain direct-read and binary surfaces without misclassifying them as durable ETL streams.
- Admin/elevated/destructive direct-read or binary operations remain human-gated.

### 4. Reverse ETL and Mutation Parity

**Goal:** Map safe writes/mutations to reverse ETL actions with approval gates.

**Depends on:** Workstream 1.

**Success criteria:**

- Product-safe mutations across REST, GraphQL, XML/SOAP, file/object, queue, database/native, and other protocol-specific operations map to reverse ETL actions or typed exclusions.
- Plan, preview, approval, and execute semantics are preserved for every write path.
- Runtime help, manual docs, and website docs consistently describe plan → preview → approval → execute for write paths.
- Destructive/admin/elevated-scope writes are human-gated and never exposed as generic raw write tools.

### 5. Conformance and Certification Enforcement

**Goal:** Make validated gates authoritative for connector parity status.

**Depends on:** Workstreams 2–4.

**Success criteria:**

- Conformance validates schemas, fixtures, surface manifests, streams, writes, direct-read/binary metadata, cursor/pagination behavior, docs, and de-duplication contracts.
- Certification reports distinguish replay/fixture success, live success, missing credentials (`uncertified`), typed blockers, and failures.
- Public or PR-facing connector parity claims derive from generated conformance/certification artifacts.

## Phase Mapping

The existing phase files are preserved in this refresh by request. They still map to the roadmap as follows:

| Existing phase | Roadmap workstream | Status |
|---|---|---|
| Phase 1: Inventory and Surface Reconciliation | Workstreams 0–1 | In progress |
| Phase 2: Durable Read and ETL Parity | Workstream 2 | Not started |
| Phase 3: Direct-Read, Binary, and Native Surface Parity | Workstream 3 | Not started |
| Phase 4: Reverse ETL and Mutation Parity | Workstream 4 | Not started |
| Phase 5: Conformance and Certification Enforcement | Workstream 5 | Not started |

## Progress

| Workstream | Status | Notes |
|---|---|---|
| 0. GSD Runtime and Agent Enablement | In progress | Official docs pinned; Pi adapter and agent guidance refreshed. |
| 1. Inventory and Surface Reconciliation | In progress | Planning exists; generated inventory still must be reviewed before fanout. |
| 2. Durable Read and ETL Parity | Not started | Blocked on Workstream 1. |
| 3. Direct-Read, Binary, and Native Surface Parity | Not started | Blocked on Workstream 1. |
| 4. Reverse ETL and Mutation Parity | Not started | Blocked on Workstream 1. |
| 5. Conformance and Certification Enforcement | Not started | Blocked on Workstreams 2–4. |

## Command Path for Future Updates

Use official GSD commands through the repo-local adapter:

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt map-codebase --fast
scripts/gsd prompt new-project --from-existing --non-interactive
scripts/gsd prompt plan-phase 1 --skip-research
scripts/gsd prompt programming-loop init --phase <phase> --dry-run
```

In Pi after trust/reload:

```text
/gsd doctor
/gsd list
/gsd map-codebase --fast
/gsd plan-phase 1 --skip-research
/gsd-programming-loop init --phase <phase> --dry-run
```

---
*Roadmap refreshed: 2026-07-08 via repo-local official GSD Core Pi adapter; `.planning/phases/**` intentionally unchanged.*
