# Native Go All Connectors PRD

## Objective

Enable the full generated connector catalog through native Go runtime bindings while preserving the existing `pm` CLI architecture, GitHub connector behavior, approval-gated reverse ETL, and secret-redaction guarantees.

## Requirements

- All 647 catalog entries are returned as `implementation_status=enabled`.
- Every catalog slug is registered as a runnable Go connector.
- Sources support fixture-backed `check`, `catalog`, `read`, ETL metadata, stream state, and conformance evidence.
- Destinations support fixture-backed `check`, `catalog`, `write`, reverse ETL approval flow, receipts, and conformance evidence.
- Database and destination runtime families expose SELECT-only query capability.
- CDC-capable database sources expose fixture-backed CDC conformance metadata and events.
- No connector image, Python, Java, Ruby, shell plugin, or untrusted dynamic runtime is executed.
- Connector docs and skills remain generated, detailed, and secret-free.

## Acceptance

- `pm connectors list --all --json` returns 647 enabled entries.
- `pm connectors inspect <slug>` renders man-style docs for any catalog slug.
- `pm etl check`, `pm etl catalog`, and `pm etl read` work for native catalog source slugs in fixture mode.
- `pm reverse plan`, `pm reverse preview`, and `pm reverse run` work for native catalog destination slugs through approval and local receipts.
- `make verify` passes.
