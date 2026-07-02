# Overview

Fleetio is a wave2 fan-out migration of `internal/connectors/fleetio` (the
hand-written legacy connector this bundle migrates; the legacy package stays
registered and unchanged until wave6's registry flip). It reads Fleetio
fleet management data — vehicles, contacts, fuel entries, issues, and
service entries — through the Fleetio REST API v1.

## Auth setup

Fleetio requires TWO secrets on every request, both mandatory (`spec.json`'s
`required: ["api_key", "account_token"]`, matching legacy's `Check`, which
errors if either is missing):

1. `api_key` — sent as `Authorization: Token <api_key>` via an
   `api_key_header` auth spec (`header: Authorization`, `prefix: "Token "`),
   byte-for-byte identical to legacy's
   `connsdk.APIKeyHeader("Authorization", apiKey, "Token ")` construction.
2. `account_token` — sent as a static `Account-Token` request header
   (`streams.json`'s `base.headers`), identical to legacy's
   `DefaultHeaders: {"Account-Token": accountToken}`.

Provide `base_url` to override the Fleetio API root (defaults to
`https://secure.fleetio.com/api/v1`, matching legacy's
`fleetioDefaultBaseURL` constant via `spec.json`'s `default`
materialization) — also the override mechanism for tests/proxies.

## Streams notes

All 5 streams (`vehicles`, `contacts`, `fuel_entries`, `issues`,
`service_entries`) share the same shape: `GET` against the Fleetio index
endpoint, records extracted from the response's top-level `records` array
(legacy's `connsdk.RecordsAt(resp.Body, "records")`), primary key `["id"]`
(Fleetio ids are integers, not strings — schema types tightened to
`"integer"` rather than a widened string union, per conventions.md's typed-
extraction rule). Each schema declares `x-cursor-field: updated_at` for
manifest-surface parity (every Fleetio object carries `updated_at`), but —
see "Known limits" below — no stream declares an `incremental` block,
matching legacy's `harvest()` exactly (it always full-refreshes with an
initial empty cursor; `InitialState` never seeds anything but an empty
string).

Pagination follows Fleetio's cursor convention (`pagination.type: cursor`
with `cursor_param: start_cursor`, `token_path: next_cursor`): the response
envelope is `{records:[...], next_cursor:"..."}`; the next page is requested
with `start_cursor=<next_cursor>`, and pagination stops when `next_cursor`
is `null`/absent — `connsdk.StringAt` reads a JSON `null` as `""`, which is
exactly the `token_path` paginator's stop condition, matching legacy's own
`next == "" || len(records) == 0` stop rule (no `stop_path` needed — there
is no separate boolean "has more" flag the way Zendesk's `has_more` works).
Every request sends `per_page` (default `100`, matching legacy's
`fleetioDefaultPageSize`) via each stream's optional `query` entry
(`omit_when_absent`-style `default`), not via `pagination.size_param`/
`page_size`, which the `cursor`+`token_path` paginator constructor never
reads (only `page_number`/`offset_limit` do — conventions.md's F6 lesson).

## Write actions & risks

None. Legacy `fleetio` is read-only (`Write` returns
`connectors.ErrUnsupportedOperation`); `metadata.json` declares
`capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- Full Fleetio API surface (work orders, inspection submissions, parts,
  vendors, expense entries, meter entries, etc.) is out of scope for wave2;
  see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass
  B capability expansion"}` entries. Only the 5 legacy-parity streams are
  implemented.
- **No server-side incremental filter (matches legacy exactly)**: legacy
  never sends any `updated_at`-style filter query param; no stream in this
  bundle declares an `incremental` block. Each schema still declares
  `x-cursor-field: updated_at` for manifest-surface documentation only.
- **`check` request has no `per_page` bound**: legacy's `Check` issues a
  bounded `GET /vehicles?per_page=1` (`fleetio.go:88-89`) specifically to
  keep the health-check response small. The engine's `HTTPBase.Check`
  (`RequestSpec`) has no `query` field at all — ANY `query` object declared
  under `streams.json`'s `base.check` is silently dropped at read time
  (`engine.Check`, `read.go`, calls `rt.Requester.Do(ctx, method, checkPath,
  nil, nil)` — the query argument is always `nil`), so this bundle's check
  request is unconditionally `GET /vehicles` with no query string at all.
  This never changes emitted record DATA (`Check` never emits records, only
  verifies connectivity/auth) and Fleetio's `/vehicles` list endpoint
  returns the same collection either way — a strictly larger, still-
  successful response body, not a different one.
- **`page_size`/`max_pages` accepted-range validation is not enforced by
  this bundle**: legacy validates `page_size` is between 1 and 100
  (`fleetioMaxPageSize`) and `max_pages` is a non-negative integer (or
  `all`/`unlimited`), rejecting an invalid value with an error before ever
  issuing a request. The declarative dialect has no config-value
  range-validation primitive; an out-of-range `page_size` here is sent to
  Fleetio as-is rather than rejected client-side. This never changes emitted
  record DATA for any legacy-VALID input (the same in-range `page_size`
  produces the identical `per_page` query value either side) — a
  client-side-validation narrowing, not a data-parity deviation. Similarly,
  legacy's `max_pages` config knob (unlimited-by-default, capped-if-set) has
  no dedicated `spec.json` property here: the engine's `PaginationSpec.MaxPages`
  hard cap is a `streams.json`-declared value, not a runtime
  `config.*`-driven override; this bundle leaves pagination unbounded on
  every stream, matching legacy's own default (`max_pages` unset ->
  unlimited).
