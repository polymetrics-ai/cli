# Overview

FireHydrant is a wave2 fan-out migration of `internal/connectors/firehydrant`
(the hand-written legacy connector this bundle migrates; the legacy package
stays registered and unchanged until wave6's registry flip). It reads
FireHydrant incidents, services, teams, environments, and functionalities
through the FireHydrant REST API v1.

## Auth setup

Provide `api_token` as a secret (FireHydrant Settings > API Keys). It is
sent as `Authorization: Bearer <api_token>`, byte-for-byte identical to
legacy's `connsdk.Bearer(secret)` construction.

Provide `base_url` to override the FireHydrant API root (defaults to
`https://api.firehydrant.io/v1`, matching legacy's `firehydrantDefaultBaseURL`
constant via `spec.json`'s `default` materialization) — also the override
mechanism for tests/proxies. Any override must be an absolute `http(s)` URL
with a host; the engine's own SSRF guards on `base.url`/path interpolation
provide the equivalent of legacy's `firehydrantBaseURL` scheme/host
validation.

## Streams notes

All 5 streams (`incidents`, `services`, `teams`, `environments`,
`functionalities`) share the same shape: `GET` against the FireHydrant list
endpoint, records extracted from the response's top-level `data` array
(legacy's `connsdk.RecordsAt(resp.Body, "data")`), primary key `["id"]`.
Each schema declares `x-cursor-field: updated_at` for manifest-surface
parity (every FireHydrant object carries `updated_at`), but — see "Known
limits" below — no stream declares an `incremental` block, matching legacy's
`harvest()` exactly (it always full-refreshes; there is no
request-side incremental filter anywhere in the legacy connector).

Pagination follows FireHydrant's numeric page convention
(`pagination.type: cursor` with `cursor_param: page`, `token_path:
pagination.next`): the response envelope is `{data:[...],
pagination:{page, next, prev}}`; `pagination.next` holds the next page
number, or `null` once the last page is reached. This maps directly onto the
engine's `cursor`+`token_path` paginator: `connsdk.StringAt` reads a JSON
`null` as `""`, which is exactly the paginator's stop condition (no
`stop_path` is declared — FireHydrant's stop signal IS the token's own
falsiness, there is no separate boolean flag the way Zendesk's `has_more`
works). Every request sends `per_page` (default `50`, matching legacy's
`firehydrantDefaultPageSize`) via each stream's optional `query` entry
(`omit_when_absent`-style `default`), not via `pagination.size_param`/
`page_size`, which the `cursor`+`token_path` paginator constructor never
reads (only `page_number`/`offset_limit` do — conventions.md's F6 lesson,
mirrors zendesk-support's identical `page[size]`-via-static-query pattern).

## Write actions & risks

None. Legacy `firehydrant` is read-only (`Write` returns
`connectors.ErrUnsupportedOperation`); `metadata.json` declares
`capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- Full FireHydrant API surface (retrospectives, runbooks, tasks, changes,
  on-call schedules, incident milestones, etc.) is out of scope for wave2;
  see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass
  B capability expansion"}` entries. Only the 5 legacy-parity streams are
  implemented.
- **No server-side incremental filter (matches legacy exactly)**: legacy's
  `harvest()` never sends any `updated_at`-style filter query param and
  `Read` always starts a page-1 full scan; no stream in this bundle declares
  an `incremental` block. Each schema still declares `x-cursor-field:
  updated_at` for manifest-surface documentation only (the
  zendesk-support/calendly precedent for a schema-only cursor-field
  annotation with no wired `incremental` block).
- **`check` request has no `per_page` bound**: legacy's `Check` issues a
  bounded `GET /environments?per_page=1` (`firehydrant.go:84`) specifically
  to keep the health-check response small. The engine's `HTTPBase.Check`
  (`RequestSpec`) has no `query` field at all — ANY `query` object declared
  under `streams.json`'s `base.check` is silently dropped at read time
  (`engine.Check`, `read.go`, calls `rt.Requester.Do(ctx, method, checkPath,
  nil, nil)` — the query argument is always `nil`), so this bundle's check
  request is unconditionally `GET /environments` with no query string at
  all. This never changes emitted record DATA (`Check` never emits records,
  only verifies connectivity/auth) and FireHydrant's `/environments` list
  endpoint returns the same collection either way — a strictly larger,
  still-successful response body, not a different one.
- **`page_size` accepted range is documented, not enforced by this bundle**:
  legacy validates `page_size` is between 1 and 200
  (`firehydrantMaxPageSize`) and rejects an out-of-range value with an
  error before ever issuing a request. The declarative dialect has no
  config-value range-validation primitive; an out-of-range `page_size` here
  is sent to FireHydrant as-is rather than rejected client-side. This never
  changes emitted record DATA for any legacy-VALID input (the same
  in-range `page_size` produces the identical `per_page` query value either
  side), so it is a client-side-validation narrowing, not a data-parity
  deviation — an operator who supplies an out-of-range value gets FireHydrant's
  own error response instead of a client-side one.
- `max_pages` (legacy's unlimited-by-default, capped-if-set page-count knob)
  has no dedicated `spec.json` property in this bundle: the engine's own
  `PaginationSpec.MaxPages` hard cap is a `streams.json`-declared value, not
  a runtime `config.*`-driven override, so legacy's `all`/`unlimited`/integer
  config surface for `max_pages` is not reproduced. This bundle leaves
  pagination unbounded (`MaxPages` unset on every stream), matching legacy's
  own default (`max_pages` unset -> unlimited) exactly; only a
  config-provided FINITE `max_pages` override is out of scope here.
