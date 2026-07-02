# Overview

VWO (Visual Website Optimizer) is an A/B testing and conversion optimization platform. This bundle
reads campaigns from the VWO API v2 (`GET {base_url}/campaigns`). It migrates
`internal/connectors/vwo` (the hand-written legacy connector), which stays registered and unchanged
until wave6's registry flip. Read-only; a single `campaigns` stream, no pagination.

## Auth setup

`api_key` (required, `x-secret`) is VWO's API token, sent as `Authorization: Token <api_key>` via
`streams.json` `base.auth`'s `api_key_header` mode (`header: "Authorization"`, `prefix: "Token "`),
matching legacy's `connsdk.APIKeyHeader("Authorization", key, "Token ")` exactly.

`base_url` defaults to `https://app.vwo.com/api/v2`, matching legacy's `defaultBaseURL` constant,
materialized via `spec.json`'s `"default"` value.

## Streams notes

`campaigns` reads `GET /campaigns` and extracts records from the response's top-level `campaigns`
array, matching legacy's `connsdk.RecordsAt(resp.Body, "campaigns")`. Legacy maps `id` from the raw
`id` field, always stringified via `fmt.Sprint` regardless of its raw JSON type (VWO's wire shape is
a bare integer, as recorded in this bundle's fixtures). This bundle reproduces that with
`computed_fields`' `"id": "{{ record.id | last_path_segment }}"` — the `last_path_segment` filter
forces string output via `Interpolate` (a bare `{{ record.id }}` reference would instead trigger
typed extraction and copy the raw JSON integer, which is not what legacy emits) while passing a
delimiter-free numeric value through unchanged. `name`/`status`/`created_at` are copied through by
plain schema projection (exact key match, no rename needed).

`start_date` is an OPTIONAL config value legacy passes straight through as a `start_date` query
parameter whenever configured (`q.Set("start_date", start)` — no incremental cursor tracking, no
state persistence, just a static request-time filter). This bundle wires the identical behavior via
`stream.Query`'s opt-in optional-query dialect: `{"template": "{{ config.start_date }}",
"omit_when_absent": true}` — sent only when `start_date` is configured, omitted entirely otherwise,
exactly like legacy's `if start := ...; start != ""` guard. `created_at` is declared as
`x-cursor-field` for catalog/manifest parity (legacy's own `Catalog` declares
`CursorFields: []string{"created_at"}`), but — matching legacy exactly — no `incremental` block is
declared on the stream: legacy never advances a persisted cursor or filters by `created_at` on
repeat syncs, so this bundle does not either (declaring an `incremental` block here would be new,
behavior-changing filtering legacy never performed).

## Write actions & risks

None. `capabilities.write` is `false`; legacy's `Write` is an unconditional
`connectors.ErrUnsupportedOperation` stub, and this bundle ships no `writes.json`.

## Known limits

- Only the `campaigns` stream is migrated (legacy's own only stream). VWO's much larger API surface
  (goals, campaign reports/insights, account settings, etc.) is intentionally out of scope for this
  wave — see `api_surface.json`.
- No pagination: legacy's `Read` issues exactly one unpaginated request per sync, so this bundle
  declares no `pagination` block, matching legacy's actual behavior exactly.
- `created_at` is catalog-declared as a cursor field (parity with legacy's `Catalog`) but is not a
  functioning incremental cursor in either legacy or this bundle — `start_date` is a plain
  passthrough filter, not a state-tracked lower bound. This is an honest reflection of legacy's own
  behavior, not a scope narrowing introduced by migration.
