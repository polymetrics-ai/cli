# Research — Connector Surface Features

**Generated via:** Upstream `/gsd:new-project --auto` research step shape  
**Date:** 2026-07-08

## Table Stakes

Connector parity requires coverage of every documented product-safe connector capability, not just REST endpoints.

### Surface Classes

1. **ETL streams** — durable record collections usable through catalog/read/ETL.
2. **Reverse ETL write actions** — product-safe mutations through plan/preview/approval/run.
3. **Direct-read commands** — useful non-durable reads that should not be modeled as sync streams.
4. **Binary transfer commands** — artifacts, archives, attachments, files, documents, exports, media.
5. **Native protocol capabilities** — database, CDC, queue, file/object, and custom system surfaces.
6. **Webhook/event/audit-log surfaces** — event resources, logs, queues, and replayable/auditable event data.
7. **Typed exclusions** — destructive, elevated-scope, deprecated, duplicate, non-data, or out-of-scope operations.

### Protocol Families to Preserve

- REST/HTTP JSON.
- GraphQL queries, mutations, and subscription/event-style surfaces where documented.
- XML/SOAP/XML feeds.
- CSV, TSV, NDJSON, and report-export protocols.
- Binary upload/download protocols.
- File/object storage APIs and signed transfer flows.
- SQL/database and CDC protocols.
- Queues, webhooks, audit logs, and event streams.

## Differentiators

- Agent-safe CLI with JSON output and deterministic exit codes.
- Reverse ETL approval flow instead of raw mutation tools.
- Conformance and certification as the source of truth for connector capability claims.
- Explicit de-duplication so generated API docs, product guides, and protocol schemas do not inflate coverage.

## Anti-Features

- Generic shell execution.
- Generic HTTP write tools.
- Generic SQL write tools.
- Silent approximation of unsupported connector behavior.
- Credentialed live tests as a requirement for no-secret planning gates.
