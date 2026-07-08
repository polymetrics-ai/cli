# Research — Connector Technology Stack

**Generated via:** Upstream `/gsd:new-project --auto` research step shape  
**Date:** 2026-07-08

## Current Stack Signals

- Go CLI monolith with `pm` binary.
- JSON Schema connector bundles.
- Declarative HTTP engine plus hook/native escape hatches.
- Local verification via `make verify` and Go test/build/vet.
- Optional runtime services and DuckDB integration.

## Connector Technology Signals from Current Repo

Working-tree scans found docs/defs mentions for multiple connector technologies:

- `graphql`: 8 connector dirs mentioning GraphQL.
- `xml`: 13 connector dirs mentioning XML.
- `soap`: 1 connector dir mentioning SOAP.
- `csv`: 35 connector dirs mentioning CSV.
- `ndjson`: 2 connector dirs mentioning NDJSON.
- `binary`: 142 connector dirs mentioning binary-related language.
- `download`: 90 connector dirs mentioning downloads.
- `multipart`: 74 connector dirs mentioning multipart.
- `queue`: 24 connector dirs mentioning queue-related language.
- Native connector dirs include database, queue, built-in, and custom protocol implementations.

These are broad text-scan signals, not final classifications. Phase 1 must generate authoritative connector-by-connector inventory.

## Documentation Sources

Use organization/repo docs as primary constraints:

- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `docs/architecture/connector-certification-design.md`
- `docs/plans/universal-programming-loop-prd.md`
- Per-connector `docs.md`, `api_surface.json`, schemas, streams, writes, fixtures.

## No New Dependencies

Planning should not add dependencies. Protocol-specific implementation dependencies are future human-gated decisions.
