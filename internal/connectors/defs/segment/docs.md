# Overview

Segment is a definition-owned connector for the official Segment Public API OpenAPI 3.0.3 surface published from docs.segmentapis.com (declared version 73.0.8). This bundle records all 197 official method/path operations exactly once in `api_surface.json` and `operations.json`: 79 ETL reads, 1 audit-events CDC/changefeed read, 19 bounded direct provider queries, 96 approval-gated reverse ETL writes, 1 blocked binary/file workflow, and 1 excluded testing endpoint.

# Auth setup

Use a Segment Public API token as `api_token`. Add it from an environment variable or stdin, never chat or shell history. The connector sends it as a Bearer token and supports a non-secret `base_url` override for tests and regional endpoints.

# Streams notes

Read streams are generated from fixed OpenAPI GET operations and use connector-relative paths only. List-shaped streams use Segment's cursor-style `pagination.count` / `pagination.cursor` query names and fixture replay with sanitized synthetic records. Nested streams require their documented path parameters as non-secret config fields. `audit_events` is the CDC/changefeed surface and declares `timestamp` as the incremental cursor.

# Write actions & risks

Every non-read mutation in the official lane is a named reverse ETL action with a fixed method and connector-relative path. Path parameters are record fields, request bodies are schema-derived, and DELETE actions require the `destructive` confirmation challenge. Execution remains the shared reverse ETL sequence: plan, preview, explicit approval token, execute. JSON plan output redacts approval tokens so agents cannot self-approve external Segment mutations.

# Known limits

No live Segment provider calls or certification artifacts are included in this fixture-only parity slice. `createDownload` is tracked as a binary/file workflow but remains blocked because this connector does not expose a generic local file download executor. `echo` is excluded as a non-data testing endpoint. OpenAPI union schemas are conservatively represented in the engine's minimal JSON Schema dialect; operation IDs, source URLs, and fixed paths remain the authoritative mapping.
