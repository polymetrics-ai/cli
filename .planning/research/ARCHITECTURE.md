# Research: Architecture

**Generated via:** `scripts/gsd prompt map-codebase --fast`.

## Architecture Hypothesis

Polymetrics is a Go-only CLI monolith whose connector layer is moving to data-driven JSON bundles interpreted by a declarative engine. Hooks and native implementations are deliberate escape hatches, not the default. Certification and conformance should provide the evidence base for connector parity.

## Key Architecture Points

- `cmd/pm` provides the CLI entrypoint.
- `internal/connectors/defs/` carries embedded connector bundles.
- `internal/connectors/engine/` interprets declarative definitions.
- `internal/connectors/hooks/` and `internal/connectors/native/` provide non-declarative coverage.
- Reverse ETL uses plan, preview, approval, execute.
- Planning and agent workflows now use `.gsd/` and `.pi/` official GSD adapter resources.

## Implication for Connector Parity

The architecture supports multi-surface coverage only if inventory correctly classifies each upstream operation into stream, write, direct-read, binary, native, hook, exclusion, or human-gated categories. Classification must precede fanout.

---
*Architecture research refreshed: 2026-07-08.*
