# Lucid ELD

## Overview

Lucid ELD uses the DriveHOS/Lucid ELD Partner API v2 OpenAPI 2.0 document at `https://api.drivehos.app/partner/swagger/doc.json` as the authoritative source. The document was fetched on 2026-07-30 with SHA-256 `1b3756f4c69c9133e24754a856d2fe9ec2b08768edd5dec25b899f564ddb7ec4` and lists exactly eight GET operations.

This Tier-1 declarative bundle covers all official GET operations as either ETL streams or bounded direct reads:

- Streams: `drivers`, `vehicles`, `vehicle_location_history`.
- Direct reads: `company info get`, `drivers get`, `vehicles get`, `latest driver statuses list`, `latest vehicle statuses list`.

## Auth setup

Every official operation requires both credential-shaped headers:

- `X-API-Provider-Key`: Provider API Key (with rate limiting).
- `X-API-Company-Key`: Company API Key (no rate limiting).

The connection spec stores these as `provider_api_key` and `company_api_key`, and both fields are marked `x-secret: true`. The bundle sends them only as headers. No OpenAPI `securityDefinitions` block exists in the official document.

`base_url` defaults to the public origin hosting the official Swagger document, `https://api.drivehos.app`; override it only if Lucid/DriveHOS supplies a different Partner API origin. `vehicle_location_history` also requires `vehicle_id`, `start_date`, and `end_date` config values. Supply them through stored connector config or explicit `--config vehicle_id=... --config start_date=... --config end_date=...` overrides. The official wire date format is `MM-DD-YYYY`; this bundle does not reformat dates to RFC3339.

## Streams notes

`drivers` and `vehicles` use page-number pagination with provider query parameters `page` and `limit`, `start_page: 1`, `page_size: 100`, and a deterministic `max_pages: 2` Tier-1 bound so replay fixtures exercise two-page pagination without unbounded test loops. The Lucid ELD command surface does not expose provider `page`, provider page-size `limit`, or vehicle `status` filters in this Tier-1 pilot; the global `--limit` flag remains a client-side ETL row cap only.

`vehicle_location_history` uses cursor pagination with request query parameter `next_page_token` and response envelope field `next_page_token`. It sends `start_date` and `end_date` from connector config plus fixed `limit=100`; `start_date` and `end_date` must already be in `MM-DD-YYYY` format. The engine resolves required `config.start_date` and `config.end_date` templates before request-query overrides, so separate command flags cannot substitute for the config values.

The latest-status direct reads expose only the optional `driver_id`/`vehicle_id` query filters. They use the fixed operation default `query.limit=100` and do not expose provider `page` or page-size overrides.

Fixture files under `fixtures/streams/**` are synthetic conformance test doubles, not captured live data and not assertions about real DriveHOS record fields. They contain plausible synthetic values only to exercise the documented response envelope, pagination, CLI, and conformance plumbing.

## Write actions & risks

The official DriveHOS/Lucid ELD Partner API v2 OpenAPI document has no documented POST, PUT, PATCH, DELETE, report-generation, webhook-registration, or binary/media operations. This connector is read-only, sets `capabilities.write` to `false`, and intentionally omits `writes.json`.

Reverse ETL remains unavailable for Lucid ELD unless a future official source documents product-safe mutations and they are modeled as typed write actions with plan, preview, approval, and execute semantics.

## Known limits

Official envelope documents `data` as untyped `{}`, no sample response available without credentials, WithTerminal is capability-only not wire-schema evidence. Therefore the three stream schemas are intentionally open object schemas with no fixed `properties`, no `x-primary-key`, and no `x-cursor-field`, and each stream uses `projection: "passthrough"` so raw records are not dropped by an empty declared-properties allowlist.

Because no primary keys are evidenced, sync-mode derivation exposes full-refresh append/overwrite modes without dedup variants. This is a tracked follow-up for #1951/#1955 once live or sample response evidence becomes available. WithTerminal field-coverage tables must not be copied into these schemas unless future DriveHOS wire evidence independently proves those field names.

The `max_pages: 2` bounds and fixed provider page sizes are deterministic Tier-1 pilot choices for local conformance and small reads, not evidence of a provider-side limit. Expand or make provider pagination/status filters operator-configurable only with a separate engine-supported design and focused CLI/runtime tests proving the flags reach the actual request without colliding with global CLI semantics.
