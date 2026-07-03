# Overview

Appfigures is a declarative-HTTP migration. It reads Appfigures app-store reviews, products, sales
and ratings report aggregates, and store categories through the read-only Appfigures v2 REST API
(`GET https://api.appfigures.com/v2/...`). This bundle targets full capability parity with
`internal/connectors/appfigures` (the hand-written connector it migrates) across all 5 legacy
streams; the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide an Appfigures Personal Access Token via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's
`connsdk.Bearer(secret)` (`appfigures.go:299`). `base_url` defaults to `https://api.appfigures.com/v2`
and may be overridden for tests/proxies.

## Streams notes

`reviews` is a `GET /reviews` list endpoint whose records live at the `reviews` body key, primary
key `["id"]`. Pagination follows Appfigures' page-number convention (`pagination.type:
page_number`, `page_param: page`, `size_param: count`, `page_size: 100`) — legacy's `readPaged`
sends `count=<page_size>&page=<n>` and stops on `this_page >= pages` or a short page; the engine's
`page_number` paginator stops purely on a short page (`recordCount < page_size`), which is
behaviorally identical for every real Appfigures response (the API always returns exactly
`count` records except on the final page) and is the same mapping already used for other
page-number-paginated bundles in this repo (e.g. `appcues`). Optional per-request filters
(`search_store` -> `store`, `group_by` -> `group_by`, `start_date` -> `start`, `end_date` -> `end`)
are wired via the opt-in optional-query dialect (`omit_when_absent: true`), matching legacy's
`appfiguresQuery` exactly: each is sent only when its config value is set, never as an empty
string.

`products`, `sales`, `ratings`, and `categories` all read a JSON object keyed by an arbitrary id
(e.g. `products/mine` returns `{"111":{...},"222":{...}}`) and are unpaginated single-request
reads (`pagination: {"type":"none"}` overrides the base's `page_number` block, matching legacy's
`readKeyedObject`, which issues exactly one request per call). Each declares
`records: {"path":"","keyed_object":true}` — the S4 engine mini-wave's keyed-object flatten
primitive (`docs/migration/conventions.md` §3) — which explodes every value of the root object
into its own record, exactly reproducing legacy's `flattenKeyedObject` (`appfigures.go:224`).
`key_field` is intentionally left unset on all four: legacy's `flattenKeyedObject` never stamps
the map key onto the record (each value object already carries its own natural id/date field —
`id` for `products`, `date` for `sales`/`ratings`, `id` for `categories`), so setting `key_field`
would add a field legacy never emits.

- `products` — `GET /products/mine`, primary key `["id"]`; `sales`/`group_by`/`store`/date
  filters do not apply to this endpoint (legacy sends none), so no `query` block is declared.
- `sales` — `GET /reports/sales`, no primary key (matches legacy's `appfiguresSalesFields`, which
  declares none — `date` alone is not guaranteed unique across products/stores in the real API).
  Same optional `store`/`group_by`/`start`/`end` filters as `reviews`.
- `ratings` — `GET /reports/ratings`, no primary key (matches legacy). Same optional filters as
  `sales`.
- `categories` — `GET /data/categories`, primary key `["id"]`; reference data, no date filters
  (matches legacy, which sends no query params for this endpoint).

None of the 5 streams are incremental — Appfigures' v2 API has no server-side cursor filter
legacy uses (no `CursorFields` declared on any legacy stream).

## Write actions & risks

None. Legacy's `Write` unconditionally returns `connectors.ErrUnsupportedOperation`;
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- `page_size`/`max_pages` config overrides legacy exposes for `reviews` (`appfiguresPageSize`/
  `appfiguresMaxPages`, clamped 1-500 / `all`/`unlimited`) are not runtime-configurable here: the
  engine's `page_number` paginator's `PageSize` is a static int set once in `streams.json`, not
  template-resolvable, and `PaginationSpec` has no `MaxPages` field read by this paginator type
  wired to a config knob. `spec.json` intentionally does not declare `page_size`/`max_pages` (a
  declared-but-unwireable key is worse than an absent one, per conventions.md F6).
- The keyed-object streams (`products`/`sales`/`ratings`/`categories`) are read in a single
  request with no bound on response size, matching legacy exactly (legacy's `readKeyedObject` has
  no pagination or size cap either — Appfigures' report endpoints return their entire result set
  in one body).
