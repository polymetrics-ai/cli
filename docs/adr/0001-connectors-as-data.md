# ADR 0001 — Connectors as immutable authoring data and rendered execution data

- Status: Superseded by the vNext connector data model (2026-09-01)
- Current architecture: [`docs/connector-canon/SOURCE-LOCK-VNEXT.md`](../connector-canon/SOURCE-LOCK-VNEXT.md)

## Decision

Connector authors record provider facts in an immutable schema-4
`source.lock.json`. The connector generator validates that lock, resolves shared
schema references into one canonical per-operation descriptor, and renders the
deterministic execution JSON consumed by the runtime.

The source lock and its evidence remain outside execution. Runtime discovery,
planning, approval, authentication, encoding, transport, DuckDB materialization,
and sync execution load only the rendered bundle. A missing or malformed
execution file fails closed; there is no alternate reader or secondary route.

Every connector declares exactly seven lanes: direct read, direct write, binary
download, binary upload, ETL, reverse ETL, and sync transport. Unsupported lanes
are explicit empty arrays. One source operation may populate more than one lane
when the provider operation genuinely serves those execution semantics.

## Consequences

- Provider evidence is reviewable without becoming runtime state.
- Rendering is reproducible and drift is checked byte-for-byte.
- Runtime-invalid bindings fail before provider I/O.
- Diagnostic reports are advisory and cannot suppress a documented command.
- Connector authors follow one workflow from source facts to executable proof.

The authoritative schema, renderer, error rules, proof matrix, and author
procedure live in the current architecture document linked above.
