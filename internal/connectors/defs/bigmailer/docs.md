# Overview

BigMailer is a wave2 fan-out declarative-HTTP migration. It reads BigMailer brands and account
users through the BigMailer REST API (`GET https://api.bigmailer.io/v1/<resource>`). This bundle
is migrated from `internal/connectors/bigmailer` (the hand-written connector); the legacy package
stays registered and unchanged until wave6's registry flip. **This is a partial migration**: three
of legacy's five streams (`contacts`, `lists`, `fields`) are brand-scoped substreams requiring a
sub-resource fan-out read the Tier-1 declarative dialect cannot express — see Known limits.

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

## Write actions & risks

None. BigMailer is read-only in legacy (no safe reverse-ETL action set is exposed);
`capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's `Write`
returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`contacts`, `lists`, and `fields` are NOT migrated in this bundle (blocked).** Legacy reads
  these three streams by first listing every brand id (`listBrandIDs`, bounded by
  `bigmailerMaxBrands = 1000`), then paginating `GET /brands/{brand_id}/<resource>` once per
  brand, stamping `brand_id` onto every emitted record (`bigmailer.go:195-214`,
  `harvestSubstream`). This is a sub-resource fan-out read (design §B.7's legitimate Tier-2
  `StreamHook` trigger, conventions.md §1) — the Tier-1 declarative `streams.json` dialect has no
  mechanism to (a) issue a preliminary list-brands request, (b) fan out a per-stream read over
  each returned id, or (c) stamp a dynamically-discovered parent id onto every child record. Per
  this wave's hard rule (JSON + docs.md only, no Go/hooks), these three streams are left
  unmigrated; the legacy connector remains authoritative for them until a follow-up wave adds a
  `hooks/bigmailer/hooks.go` `StreamHook`. Blocker: `ENGINE_GAP` — see the structured result's
  `blockers[]` for this connector.
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
