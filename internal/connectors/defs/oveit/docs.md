# Overview

Oveit is an event ticketing/access-control platform. This bundle is a Tier-1 declarative migration
of `internal/connectors/oveit` (legacy Go package, read-only): it reads `events`, `orders`, and
`attendees` from the documented Oveit API (`GET https://oveit.com/api/{events,orders,attendees}`).
The legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Legacy requires both `email` (a plain config value, not a secret) and `password` (a secret),
combined into HTTP Basic auth (`Authorization: Basic base64(email:password)`). This bundle wires
the identical shape via `streams.json` `base.auth`: `{"mode":"basic","username":"{{ config.email
}}","password":"{{ secrets.password }}"}`. Both `email` and `password` are `required` in
`spec.json`, matching legacy's hard error when either is missing
(`oveit.go:147-149`, `"oveit connector requires config email and secret password"`).

`base_url` defaults to `https://oveit.com/api` (legacy's `defaultBaseURL`), materialized via
`spec.json`'s `"default"` mechanism — an unset `base_url` now round-trips to the same default
legacy applied in code.

## Streams notes

All three streams (`events`, `orders`, `attendees`) share an identical record shape and pagination
behavior, matching legacy's single `streamEndpoint`/`record()` mapping applied uniformly across all
three endpoints. Records are extracted from the top-level `data` array. Primary key is `id`; there
is no incremental cursor (legacy never filters or advances reads by a timestamp field — every read
is a full stream read), so no `x-cursor-field`/`incremental` block is declared.

Pagination is `cursor` (`token_path`) reading the next page number from the response body's
`next_page` field, matching legacy's own `harvest` loop (`oveit.go:94-121`): the cursor param name
is `page`, and legacy stops as soon as `next_page` is absent/empty. `per_page` is sent on every
request via `config.page_size` (default 100, legacy's `defaultPageSize`); legacy additionally
enforces a hard max of 500, which is documentation-only in this bundle (see Known limits).

## Write actions & risks

None. Oveit's legacy connector is read-only (`Write` always returns `ErrUnsupportedOperation`);
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **First-page request omits the `page` query param entirely.** Legacy's `harvest` loop explicitly
  sends `page=1` on its first request. The engine's `cursor`(`token_path`) paginator
  (`tokenPathCursor.Start()`) always issues the first request with no cursor param at all, only
  adding `page=<token>` once a `next_page` value is read back from a prior response. This is a
  request-shape difference only, never an emitted-record-data difference (Oveit's API, like any
  standard paginated REST list endpoint, treats an absent page param as "first page" — legacy's own
  fixture/test harness never distinguishes an explicit `page=1` from an absent one). Documented per
  the parity-deviation meta-rule (`docs/migration/conventions.md` §5); acceptable since it cannot
  change accepted-input behavior.
- **`page_size` upper bound (500) is not enforced by this bundle.** Legacy's `pageSize` helper
  rejects a `page_size` config value outside `[1, 500]` with a hard config error
  (`oveit.go:182-192`). The engine's declarative config layer has no numeric-range validation
  primitive for `spec.json` properties; `connectorgen validate`'s schema-shape checks cover type,
  not bounds. Not modeled as a value constraint; documented as scope narrowing since an
  out-of-range `page_size` value is a caller-configuration error, not something an emitted record
  would ever reflect. A caller supplying a wildly out-of-range value would simply have it forwarded
  verbatim to Oveit's `per_page` query param and rely on Oveit's own server-side clamping/error
  response.
- **`max_pages` "all"/"unlimited" string aliases are not modeled.** Legacy's `maxPages` helper
  accepts the literal strings `"all"`/`"unlimited"` (case-insensitive) as synonyms for "0 = no cap"
  in addition to an empty value. This bundle has no `max_pages` config property at all (pagination
  is capped only by the API's own `next_page` exhaustion, matching the common case of an unset
  `max_pages` in legacy) — a caller needing a hard page-count cap can rely on `MaxPages`-equivalent
  behavior once/if this bundle adds a `max_pages` spec property in a later Pass B increment.
