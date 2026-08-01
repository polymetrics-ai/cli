# Overview

Square connector parity was refreshed against the official Square OpenAPI 3 latest document (`info.version` 2.0) from developer.squareup.com on 2026-08-01. The latest Square-Version list entry at review time was `2026-07-15`.

Implemented fixture-backed surfaces:

- Streams: 110 GET operations.
- Bounded direct reads/searches: 31 POST search/bulk-retrieve/calculate/retrieve operations through fixed `operations.json` entries and `json_redacted` output policy.
- Reverse ETL writes: 153 typed actions in `writes.json`; execution remains the existing plan -> preview -> approval -> execute flow.

The complete operation ledger is `api_surface.json` and contains all 346 official operations exactly once. Blocked rows remain blocked by default and carry official-source or shared-runtime dependency evidence.

## Auth setup

Connection fields:

- `api_key` (required, secret): Square access token or OAuth access token. It is used only for Bearer authentication and is never logged.
- `base_url` (optional, default `https://connect.squareup.com`): set to `https://connect.squareupsandbox.com` for sandbox fixtures or approved live checks.
- `start_date` (optional RFC3339): lower bound for incremental payment/refund reads.

Requests send the `Square-Version: 2026-07-15` header. Connection checks call `GET /v2/locations`.

## Streams notes

Streams are generated from official GET operations and are fixture-backed under `fixtures/streams/<stream>/page_1.json`. Parameterized streams use explicit connection spec fields for their path parameters; no raw path or query escape hatch is exposed. Cursor pagination is bounded (`max_pages: 10`) and list streams include `limit=100` when the official operation documents a `limit` query parameter.

Historical high-value streams keep their stable names: `payments`, `refunds`, `customers`, and `locations`. Other streams use snake_case operation IDs.

## Write actions & risks

`writes.json` declares 153 typed Square reverse-ETL actions. Each action has a closed root `record_schema`; path parameters are explicit `path_fields`; DELETE actions use `body_type: none`, `confirm: destructive`, and idempotent 404 handling. Other destructive-style actions (cancel/disable/refund/capture/pay/void/etc.) carry `confirm: destructive` and path-field redaction when they put identifiers in URLs.

No action executes during connector inspection, validation, conformance, or docs generation. Runtime execution remains plan -> preview -> explicit approval -> execute. No arbitrary method/path/body/raw query/shell/file passthrough is exposed.

## Known limits

Blocked/planned operation counts in this fixture-only parity slice:

- Deprecated official operations: 25 (official OpenAPI `deprecated: true` / `x-deprecation`).
- OAuth/mobile authorization lifecycle: 6 (credential/token exchange and revocation are outside the connector data-plane contract).
- Generic V1 batch passthrough: 1 (would permit arbitrary nested API requests).
- Webhook/event/changefeed lifecycle: 15 (requires shared CDC foundations #2986/#2988).
- Multipart/file/attachment operations: 5 (requires a reviewed bounded binary/file executor and redaction policy).

Fixture-only parity is not Square certification. `certification.json` intentionally declares no live write pairings, direct-read candidates, or binary candidates.
