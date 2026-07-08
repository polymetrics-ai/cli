# Research Summary — Connector Parity Rebootstrap

**Generated via:** Upstream `/gsd:new-project --auto` research synthesis shape  
**Date:** 2026-07-08

## Key Finding

Connector parity must be planned as multi-surface product parity, not only REST API endpoint parity. The repo already contains evidence of REST, GraphQL, XML/SOAP, CSV/NDJSON, binary transfer, file/object storage, SQL/CDC, queue/event/webhook/audit-log, direct-read, and native connector surfaces.

## Stack Additions

No stack additions are needed for issue #122. Future implementation dependencies are human-gated.

## Table Stakes

- Generate current inventory before fanout.
- Canonicalize and de-duplicate documented operations.
- Classify each operation exactly once as ETL stream, reverse ETL write, direct-read, binary transfer, native protocol, event/webhook/audit surface, or typed exclusion.
- Preserve safety: no secrets, no generic raw tools, reverse ETL approval flow, human gates.
- Make conformance and certification the source of truth for parity claims.

## Watch Out For

- REST-only blind spots.
- Duplicated operations across multiple docs/spec sources.
- Unsafe admin/destructive writes.
- Misclassifying binary/direct-read/native surfaces as streams.
- Stale counts from old planning artifacts.

## Roadmap Implication

The first roadmap phase must be inventory and surface reconciliation. Connector fanout comes only after that inventory is generated and reviewed.
