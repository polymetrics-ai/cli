# Overview

BigMailer is a wave2 fan-out declarative-HTTP migration. It reads BigMailer brands, account users,
and brand-scoped contacts, lists, and custom fields through the BigMailer REST API
(`GET https://api.bigmailer.io/v1/<resource>`). This bundle is migrated from
`internal/connectors/bigmailer` (the hand-written connector); the legacy package stays registered
and unchanged until wave6's registry flip. All 5 legacy streams are now migrated: the 2 top-level
collections (`brands`, `users`) plus the 3 brand-scoped substreams (`contacts`, `lists`, `fields`),
which use the engine's `fan_out` dialect (S4 engine mini-wave item 2) — the `ENGINE_GAP` that
previously blocked these three streams is closed.

## Auth setup

Provide a BigMailer API key via the `api_key` secret; it is sent as the `X-API-Key` header
(`streams.json` `base.auth`'s `api_key_header` mode), matching legacy's
`connsdk.APIKeyHeader(bigmailerAuthHeader, secret, "")` (`bigmailer.go:311`). Never logged.
`base_url` defaults to `https://api.bigmailer.io/v1` and may be overridden for tests/proxies.

## Streams notes

`brands` (`GET /brands`) and `users` (`GET /users`) are top-level collections read directly;
records live at the `data` key, matching legacy's `connsdk.RecordsAt(resp.Body, "data")`. Neither
stream is incremental — legacy declares no `CursorFields` for either (BigMailer's list API supports
only cursor pagination, not a time-based incremental filter), and neither schema declares
`x-cursor-field`.

Pagination is `cursor` (`token_path: cursor`, `stop_path: has_more`): the next page is requested
with `cursor=<value>` from the response body's `cursor` field, and pagination stops when
`has_more` is not literally `"true"` OR the returned `cursor` is empty — reproducing legacy's exact
`hasMore != "true" || strings.TrimSpace(next) == ""` stop rule (`bigmailer.go:187`) via the
engine's `stop_path`-on-`tokenPathCursor` mechanism (conventions.md §3).

`contacts`, `lists`, and `fields` are brand-scoped substreams: legacy's `harvestSubstream`
(`bigmailer.go:195-214`) first lists every brand id (`listBrandIDs`, bounded defensively by
`bigmailerMaxBrands = 1000`), then paginates `GET /brands/{brand_id}/<resource>` once per brand,
stamping `brand_id` onto every emitted record. This bundle expresses the identical sequence via
`streams.json`'s `fan_out` block: `ids_from.request` issues a preliminary `GET /brands` (paginated
to exhaustion using the SAME base `cursor` pagination spec every other stream uses — the
id-listing request declares no pagination block of its own, conventions.md §3), extracts `id` off
every returned brand record, then `into.path_var: "brand_id"` threads each resolved id into
`/brands/{{ fanout.id }}/<resource>`'s path, and `stamp_field: "brand_id"` writes it onto every
emitted record after projection — matching legacy's stamped `brand_id` field exactly. The one
documented, non-blocking divergence: legacy caps the brand-id fan-out at `bigmailerMaxBrands =
1000` as a defensive bound against a runaway crawl; the engine's `fan_out.ids_from.request` has no
equivalent cap (only `PaginationSpec.MaxPages`, applied per sub-sequence, not to the id-listing
request) — an account with more than 1000 brands would fan out over all of them here versus being
capped at 1000 in legacy. This is accepted as a parity deviation (§5): it never changes emitted
record DATA for any account legacy itself would fully sync (mid-cap accounts are identical;
over-cap accounts get MORE data here, never less or wrong), and no such account exists in the
fixture/conformance surface.

## Write actions & risks

None. BigMailer is read-only in legacy (no safe reverse-ETL action set is exposed);
`capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's `Write`
returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`contacts`/`lists`/`fields` fan-out has no brand-count cap.** Legacy's `listBrandIDs` bounds
  the brand-id fan-out at `bigmailerMaxBrands = 1000` as a defensive measure. The engine's
  `fan_out.ids_from.request` fully paginates the id-listing request to exhaustion with no
  equivalent cap. Documented parity deviation (§5, ACCEPTABLE): never changes emitted data for any
  account legacy itself would fully sync; only affects the hypothetical >1000-brand account, where
  this bundle emits strictly more (never wrong or missing) data than legacy's capped crawl.
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size`
  (1-100, default 100) and `max_pages` (0/all/unlimited default) as config-driven overrides
  (`bigmailerPageSize`/`bigmailerMaxPages`, `bigmailer.go:344-372`). The engine's `cursor`
  paginator has no page-size-equivalent knob at all (BigMailer's `limit` query param is a static
  per-stream `query` literal here, matching stripe's `limit=100` static-query precedent), and
  `PaginationSpec.MaxPages` is a fixed bundle-time int, never `config.*`-templated
  (`docs/migration/conventions.md`'s searxng/bitly precedent). `limit=100` (legacy's own default)
  is baked into each stream's static `query`; neither `page_size` nor `max_pages` is declared in
  `spec.json` (F6, REVIEW.md).
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only
  reached when `config.mode == "fixture"`) is a synthetic, non-live-shape fixture generator, not a
  representation of BigMailer's real wire format for `brands`/`users` — this bundle's schemas and
  fixtures instead use the real `{data:[...], has_more, cursor}` envelope legacy's live `harvest`
  path actually decodes.
