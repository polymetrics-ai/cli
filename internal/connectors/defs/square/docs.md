# Overview

Square reads payments, refunds, customers, and locations through the Square Connect v2 REST API
(`https://connect.squareup.com/v2`). This bundle migrates `internal/connectors/square` (the
hand-written legacy connector) at capability parity; the legacy package stays registered and
unchanged until wave6's registry flip. Square is read-only here — legacy has no reverse-ETL write
surface, so `capabilities.write` is `false` and no `writes.json` is shipped.

## Auth setup

Provide a Square access token (an API key or an OAuth access token; both are bearer-shaped tokens
Square accepts interchangeably) via the `api_key` secret. It is used only for
`Authorization: Bearer <api_key>` and is never logged. Every request also sends a fixed
`Square-Version: 2024-01-18` header (pinning the dated API's response shape), matching legacy's
`squareAPIVersion` constant exactly.

## Streams notes

All 4 streams (`payments`, `refunds`, `customers`, `locations`) share Square's cursor pagination
convention: the response body carries a top-level `cursor` field; the next page is requested with
`?cursor=<value>`, and pagination stops when `cursor` is absent or empty
(`pagination.type: cursor`, `token_path: cursor`, `cursor_param: cursor` — no `stop_path`, matching
legacy's exact stop-on-empty-cursor-only behavior since Square never sends a separate boolean stop
signal). Every request sends `limit=100` (matches legacy's default `page_size`).

`payments` and `refunds` additionally send `begin_time` computed from the incremental lower bound
(the sync's persisted cursor, or the RFC3339/date `start_date` config value on a fresh sync) via the
opt-in optional-query dialect (`omit_when_absent: true`) — present only when the lower bound
resolves, absent on read paths where it does not, exactly matching legacy's
`if endpoint.timeParam != "" && beginTime != "" { base.Set(endpoint.timeParam, beginTime) }` guard.
Their schemas declare `x-cursor-field: updated_at`, matching legacy's published `CursorFields`.

`customers` and `locations` are full-refresh only: legacy's own `streamEndpoint` table sets
`timeParam: ""` for `customers` (locations never had a time param either), so the computed
incremental lower bound is silently discarded and every read re-emits the complete object set —
this bundle reproduces that exactly by declaring no `incremental` block for either stream (matching
legacy's real behavior over its catalog-only advertised `CursorFields: ["updated_at"]` for
customers, which legacy never actually wires into a request).

## Write actions & risks

None. Square is read-only in legacy (`Capabilities.Write` is `false`); `Write` always returns
`connectors.ErrUnsupportedOperation`. No `writes.json` is shipped for this bundle.

## Known limits

- Legacy accepts several secret key aliases for the same bearer token
  (`credentials.api_key`/`api_key`/`access_token`/`credentials.access_token`) since the Square
  catalog flattens an OAuth-vs-API-key `oneOf` into dotted keys and legacy takes the first non-empty
  one. This bundle declares a single canonical secret, `api_key` — the caller supplies whichever
  Square credential they have under that one key. This narrows the accepted config surface (a
  caller who only ever populated `credentials.access_token` must remap it to `api_key`) but never
  changes emitted record data for any accepted credential.
- Legacy's `is_sandbox` config flag selects the sandbox host
  (`https://connect.squareupsandbox.com/v2`) only when `base_url` is unset — a derived default (the
  base URL is a function of another config value), which the engine's `spec.json` `"default"`
  materialization mechanism (a fixed literal only) cannot express. `is_sandbox` is dropped from
  `spec.json`; sandbox testing sets `base_url` explicitly instead. Documented config-surface
  narrowing, never a silent behavior change for any caller who already sets `base_url` directly.
- Full Square API surface (orders, catalog, invoices, team members, disputes) is out of scope for
  this wave; see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting for
  Square, so none is added here (matching legacy's real, lack-of, throttling behavior).
