# Structure

**Analysis Date:** 2026-07-08
**Generated via:** Upstream `/gsd:map-codebase` workflow shape, issue #122 prompt.

## Top-Level Layout

| Path | Purpose |
|---|---|
| `cmd/pm` | Main CLI executable entry point. |
| `cmd/connectorgen` | Connector bundle validation/generation/scaffolding. |
| `cmd/iconregistrygen` | Connector icon registry generation. |
| `internal/cli` | CLI command implementations, docs/help rendering, certification CLI. |
| `internal/app` | ETL, reverse ETL, query, warehouse, connection, and sync-mode application logic. |
| `internal/connectors` | Connector interfaces, declarative engine, defs, hooks, natives, conformance, certify. |
| `internal/runtime` / `internal/worker` | Optional runtime-backed execution. |
| `internal/flow` | Flow execution surfaces. |
| `internal/schedule` | Local scheduling/crontab integration. |
| `internal/vault` | Credential storage and secret handling. |
| `docs/architecture` | Architecture and certification design docs. |
| `docs/migration` | Connector migration conventions, inventory/status, quarantine, rosters. |
| `.agents/agentic-delivery` | Issue-first delivery contracts and automated review workflows. |
| `.planning` | Active upstream GSD Core planning artifacts for issue #122 onward. |

## Connector Directory Structure

| Path | Purpose |
|---|---|
| `internal/connectors/defs/<name>/metadata.json` | Connector identity, capabilities, risk, docs URL. |
| `internal/connectors/defs/<name>/spec.json` | Connection JSON Schema with `x-secret` fields. |
| `internal/connectors/defs/<name>/streams.json` | Declarative stream/read definitions. |
| `internal/connectors/defs/<name>/writes.json` | Declarative write actions, when present. |
| `internal/connectors/defs/<name>/api_surface.json` | Documented operation coverage and exclusions. |
| `internal/connectors/defs/<name>/schemas/*.json` | Per-stream schemas with primary key/cursor extensions. |
| `internal/connectors/defs/<name>/fixtures/**` | Recorded fixture pages and write request shapes. |
| `internal/connectors/hooks/<name>/` | Connector-specific Tier 2 hooks. |
| `internal/connectors/native/<name>/` | Tier 3 native implementations. |

## Generated / Shared Files to Treat Carefully

- Hook/native generated wiring is orchestrator-owned when regenerated.
- `go.mod` / `go.sum` must not change without human approval.
- `cmd/` and `internal/` must not change in issue #122.
- Active `.planning/` is tracked; local archive paths outside `.planning/` are not active GSD context.

## Current Inventory Snapshot Inputs

During onboarding, generated shell checks observed:

- `internal/connectors/defs`: 547 connector definition directories.
- `internal/connectors/hooks`: 78 hook directories.
- `internal/connectors/native`: 37 native directories.
- `internal/connectors/defs/*/api_surface.json`: 547 files.
- `api_surface.json` endpoint rows: 29,123 total, 13,468 covered, 15,655 excluded.

These are planning inputs, not final reconciled truth; Phase 1 must regenerate and review the authoritative inventory.

---
*Structure analysis: 2026-07-08*
