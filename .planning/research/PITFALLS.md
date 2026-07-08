# Research — Pitfalls

**Generated via:** Upstream `/gsd:new-project --auto` research step shape  
**Date:** 2026-07-08

## REST-Only Planning Pitfall

Treating connector parity as only REST endpoint parity misses GraphQL, XML/SOAP, CSV/NDJSON exports, binary transfers, file/object APIs, SQL/CDC, queues/events/webhooks, and native protocols. This creates false completion and misclassified CLI surfaces.

## Duplicate Operation Pitfall

The same upstream capability can appear in OpenAPI specs, generated docs, human guides, SDK examples, GraphQL schemas, and webhook docs. Without canonical operation identity, planning can double-count the same operation as multiple streams/actions/commands.

## Unsafe Write Pitfall

Some writes are destructive, admin-only, billing-impacting, elevated-scope, or credential-management operations. These must not be normalized into routine reverse ETL actions without human approval.

## Wrong Abstraction Pitfall

Binary downloads, report exports, direct status reads, SQL queries, queue receives, and event logs should not all be forced into durable ETL streams. Correct classification matters for CLI UX, safety, conformance, and certification.

## Stale Inventory Pitfall

Prior `.planning/` counts were produced by earlier branches and phases. Phase 1 must regenerate counts from the current tree and current upstream docs.

## Secret Handling Pitfall

Planning and replay gates must not require secrets. Live certification is credential-gated and missing credentials should produce `uncertified`, not failure.
