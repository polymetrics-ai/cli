# Polymetrics CLI Connector Parity

## What This Is

Polymetrics CLI (`pm`) is a local-first Go CLI for ETL, reverse ETL, connector inspection, credential management, local warehouse queries, scheduling, and optional runtime-backed execution. This GSD Core brownfield project tracks connector parity across every documented connector surface, not only REST APIs: REST/JSON, GraphQL, XML/SOAP, CSV/NDJSON exports, binary transfers, file/object storage, SQL/CDC, queues/events/webhooks, native protocols, direct-read commands, and product-safe writes.

## Core Value

Users and agents can trust `pm` as a connector-complete, safety-gated ETL and reverse ETL interface whose advertised connector capabilities match documented upstream product surfaces without duplicate or unsafe exposure.

## Requirements

### Validated

- ✓ Single-binary Go CLI architecture exists with ETL, query, reverse ETL, scheduling, docs, and smoke verification.
- ✓ Connector Architecture v2 exists with embedded JSON definition bundles, declarative engine, hooks, native connectors, conformance, and `connectorgen validate`.
- ✓ Reverse ETL uses plan, preview, approval, and execute semantics; raw generic shell, generic HTTP write, and generic SQL write tools remain out of scope.
- ✓ Certification harness components exist under `internal/connectors/certify/` with source, write, replay, report, ledger, and sweeper stages.

### Active

- [ ] Reconcile the current connector inventory before any connector fanout.
- [ ] Preserve connector parity as the north-star milestone across GSD planning artifacts.
- [ ] Classify all documented connector surfaces across protocols without duplicating operations.
- [ ] Ensure every product-safe documented stream/report/event/log is available through `pm` CLI and ETL or a typed exclusion.
- [ ] Ensure every product-safe documented write/mutation is available through reverse ETL plan/preview/approval/run or a typed exclusion.
- [ ] Ensure direct-read and binary-transfer endpoints are supported where product-safe and human-gated where risky.
- [ ] Ensure conformance and certification gates enforce parity before rollout claims.

### Out of Scope

- Runtime connector changes for issue #122 — this issue only reboots planning state.
- New dependencies — dependency additions require explicit human approval.
- Credentialed live connector checks — no secrets are needed for planning rebootstrap.
- Destructive/admin execution — remains human-gated and outside automated planning edits.
- Generic shell, generic HTTP write, or generic SQL write tools — explicitly disallowed safety boundary.

## Context

This repository previously carried a custom/legacy `.planning/` tree with stale phase artifacts and connector counts. Issue #122 reboots active planning using upstream GSD Core workflow shape while preserving Polymetrics-specific connector parity overlays from `AGENTS.md`, `docs/migration/HANDOFF-CODEX.md`, `docs/migration/conventions.md`, `docs/architecture/connector-architecture-v2-design.md`, `docs/architecture/connector-certification-design.md`, and `docs/plans/universal-programming-loop-prd.md`.

The current codebase has a completed Connector Architecture v2 baseline on `main`: JSON connector definitions under `internal/connectors/defs/`, Tier 2 hooks, Tier 3 natives, conformance, certification, and code generation. Onboarding scans observed 547 connector definition directories, 78 hook directories, 37 native directories, and 547 `api_surface.json` files with 29,123 endpoint rows. These are initial inputs only; the first phase must regenerate and reconcile authoritative inventory before fanout.

## Constraints

- **Safety**: No secrets, credentialed checks, reverse ETL execution, destructive/admin actions, or production deploys during planning work.
- **Source scope**: Planning-only issue; `cmd/` and `internal/` must not change.
- **Human gates**: New dependencies, auth scope changes, destructive/admin operations, quality-gate reductions, and merges to `main` require human approval.
- **Architecture**: Connector runtime remains declarative-first with Tier 2 hooks and Tier 3 natives only when justified by `docs/migration/conventions.md`.
- **Surface coverage**: Connector parity covers REST, GraphQL, XML/SOAP, report exports, binary transfers, file/object storage, SQL/CDC, queues/events/webhooks, native protocols, direct-read, and writes.
- **De-duplication**: One upstream operation maps to exactly one primary classification; aliases and duplicate docs references are cross-links, not duplicated work.
- **Review**: Issue-to-PR delivery follows `.agents/agentic-delivery/contracts/issue-agent-contract.md`; PR targets `main` and uses `Closes #122`.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Rebootstrap active `.planning/` from upstream GSD Core workflow shape | The old tree was custom/legacy and should not pollute future onboarding | — Pending |
| Archive old `.planning/` outside active planning context | Preserves auditability without tracking stale artifacts | ✓ Good |
| Inventory reconciliation precedes connector fanout | Current bundle/API/certification/surface counts must be trusted before parallel work | — Pending |
| Treat connector parity as multi-surface, not REST-only | GraphQL, XML/SOAP, CSV/NDJSON, binary, file, database, CDC, queue, webhook, and direct-read surfaces exist in the repo | — Pending |
| Keep active `.planning/` tracked | Issue explicitly requires upstream-generated planning to remain versioned | ✓ Good |

---
*Last updated: 2026-07-08 after issue #122 upstream GSD Core rebootstrap onboarding*
