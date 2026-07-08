# Architecture

**Analysis Date:** 2026-07-08
**Generated via:** Upstream `/gsd:map-codebase` workflow shape, issue #122 prompt.

## Pattern Overview

**Overall:** Go-only local-first CLI monolith with declarative connector runtime and optional native/hook escape hatches.

**Key Characteristics:**
- Single executable `pm` built from `cmd/pm`.
- CLI layer delegates to application services and connector registry/runtime packages.
- Connector definitions are embedded data under `internal/connectors/defs` and interpreted by `internal/connectors/engine`.
- Safety-gated reverse ETL uses plan, preview, approval, and execute flow.
- Conformance and certification are intended to make connector capability claims testable.

## Layers

**CLI layer:**
- Purpose: Parse commands, flags, JSON output, docs/help, and user-facing errors.
- Contains: `cmd/pm`, `internal/cli`.
- Depends on: app services, connector registry, runtime helpers, safety redaction.

**Application layer:**
- Purpose: ETL, reverse ETL, query, flow, schedule, runtime-backed execution, state, warehouse, and approval flows.
- Contains: `internal/app`, `internal/flow`, `internal/schedule`, `internal/runtime`, `internal/state`, `internal/vault`.
- Depends on: connector interfaces, local state/warehouse, safety utilities.

**Connector layer:**
- Purpose: Connector definitions, runtime execution, hooks/native overrides, conformance, certification.
- Contains: `internal/connectors/defs`, `engine`, `hooks`, `native`, `connsdk`, `conformance`, `certify`, `bundleregistry`.
- Depends on: connector schemas/fixtures, HTTP helpers, native clients, hook registration.

**Planning and migration layer:**
- Purpose: Authoring conventions, migration inventories, parity status, issue-first delivery rules.
- Contains: `docs/migration`, `docs/architecture`, `.agents/`, `.planning/`.

## Data Flow

**ETL read flow:**
1. User or agent invokes `pm etl ...` / connector read surface.
2. CLI loads project config and credential references.
3. App resolves connector, stream, sync mode, state, and destination.
4. Connector runtime reads from engine/hook/native implementation.
5. Records flow to local warehouse/output; cursor/state advances.

**Reverse ETL write flow:**
1. User creates a reverse plan from warehouse table to connector/action mapping.
2. Preview validates mapping and records.
3. Approval token gates execution.
4. Writer executes product-specific write action.
5. Result/ledger/status is stored; destructive/admin paths require extra human gates.

**Connector certification flow:**
1. `pm connectors certify` stages inspect metadata, credentials, catalog, source reads, write pairings, replay, flow, schedule, redaction, and report outputs.
2. Replay/fixture gates run without secrets.
3. Live checks require explicit credentials and treat missing credentials as uncertified.

## Connector Surface Model

Planning must model a connector as a collection of surfaces, not as a REST-only API:

- ETL streams for durable record collections.
- Reverse ETL write actions for product-safe mutations.
- Direct-read commands for useful reads that are not durable streams.
- Binary transfer commands for archives, artifacts, documents, files, exports, and media.
- Native protocol capabilities for SQL/database/CDC/queue/file systems.
- GraphQL, XML/SOAP, CSV/NDJSON, webhook/event/audit-log and other protocol-specific surfaces.
- Typed exclusions for destructive, elevated-scope, deprecated, duplicate, non-data, or out-of-scope operations.

## De-duplication Rule

A documented upstream operation must have exactly one primary classification. Duplicate docs pages, aliases, generated API references, and product guide variants should point to the same canonical operation identity instead of creating duplicate work.

---
*Architecture analysis: 2026-07-08*
