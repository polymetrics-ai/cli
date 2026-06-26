# Architecture Hardening Summary

## Goal

Implement the seven architecture/safety improvements identified in the repo assessment using GSD programming-loop discipline, TDD slices, and multiple subagents.

## Completed

- Enforced reverse ETL approval tokens as single-use and bound execution to the planned mapped records.
- Added central error redaction and applied it to CLI output, persisted ETL errors, persisted reverse ETL errors, and connsdk HTTP errors.
- Replaced silent CLI parsing fallback with validation errors for malformed `--config`, `--map`, and integer flags.
- Added compatibility-preserving local path policy: external warehouse/outbox write paths require `allow_external_path=true`; file source paths remain compatible.
- Split production live registry from staged connector registry: `registryset.New()` is live-only; `registryset.NewStaged()` includes all self-registered packages.
- Added `internal/runtime` with Dragonfly/Postgres adapters over existing dependencies only.
- Added `internal/state` locked atomic JSON store and wired it into `app.App` state load/save.
- Replaced the warehouse destination name check with a `connectors.LocalWarehouseMaterializer` seam.
- Added early-stop read limits through a connector-level sentinel and helper.
- Streamed deduped raw materialization into the best-record map instead of first loading all raw rows.
- Added `internal/connectors/httpsource` and migrated Google Web Fonts as a representative HTTP source template consumer.
- Regenerated the connector registry and docs after changing live/staged registry behavior.

## Subagents

- `internal/runtime` Module: implemented by subagent with tests.
- `internal/state` Module: implemented by subagent with tests.
- `internal/connectors/httpsource` Module: implemented by subagent with tests.
- Shared app/CLI/registry/materialization integration: implemented in main orchestration to avoid overlapping edits.
