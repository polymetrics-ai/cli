# Overview

RD Station Marketing is a wave2 fan-out declarative-HTTP migration. It reads RD Station Marketing
contacts, segmentations, events, landing pages, and email templates through the platform REST API
(`GET https://api.rd.services/platform/...`). This bundle migrates
`internal/connectors/rd-station-marketing` (the hand-written connector, Go package
`rdstationmarketing`); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide an RD Station Marketing OAuth access token via the `access_token` secret; it is sent as a
Bearer token (`Authorization: Bearer <access_token>`) and is never logged, matching legacy's
`connsdk.Bearer(secret)` (`rd_station_marketing.go:156`). `base_url` defaults to
`https://api.rd.services/platform` and may be overridden for tests/proxies.

## Streams notes

All 5 streams share the same shape: `GET` against the RD Station list endpoint, records at the
key matching the stream name (`contacts`/`segmentations`/`events`/`landing_pages`/
`email_templates`). Pagination is `page_number` (`page`/`page_size` query params, base default
`page_size: 125` matching legacy's `rdDefaultPageSize`/`rdMaxPageSize`, both fixed at 125 — legacy
allows a config-driven `page_size` override in the 1-125 range, but always defaults to and never
exceeds 125) — a page shorter than `page_size` is the last page. All 5 streams use the base default
(`page_size: 125`); `contacts` ships the required 2-page conformance fixture by returning a full
125-record first page (triggering the paginator to fetch page 2) and a short second page, per
`docs/migration/conventions.md` §4 — the live page size is never shrunk to make the fixture small.
The other 4 streams ship single-page fixtures.

Legacy's own halt condition additionally reads a `pagination.next_page` value from the response
body and stops early when it is empty or does not advance (`rd_station_marketing.go:118-129`),
rather than relying on a short page. The engine's `page_number` paginator has no
body-driven-next-page-token mechanism (that shape is `pagination.type: cursor` with `token_path`,
which sends the token back as a query parameter, not as a page-number/count-driven cursor) — it
stops purely on a short page. For every input where RD Station's own `page_size` items are
actually returned per page (the common case), both sides terminate identically; see Known limits.

`contacts` and `events` publish `incremental.cursor_field` (`updated_at`/`created_at`) with no
`request_param`, matching legacy's identical `CursorFields` catalog declaration with no
server-side filter — every read is a full refresh on both sides.

## Write actions & risks

None. RD Station Marketing's legacy connector implements no writes (`Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **Same-or-alternate-key field fallbacks are not modeled.** Legacy's `contactRecord` falls back
  `id` from `uuid` → `id` (`rd_station_marketing.go:188`); `namedRecord` (used by
  `segmentations`/`landing_pages`/`email_templates`) falls back `id` from `uuid` → `id` and `name`
  from `name` → `title` (`rd_station_marketing.go:190-191`); `eventRecord` falls back `id` from
  `uuid` → `id` → `event_identifier` and `event_type` from `event_type` → `event`
  (`rd_station_marketing.go:193-194`). The engine's `computed_fields` dialect has no
  coalesce/fallback filter — an `ENGINE_GAP` for expressing a same-or-alternate-key fallback
  declaratively (conventions.md §5's agilecrm/searxng precedent for this exact class of gap). Only
  the PRIMARY key name in each fallback chain (`uuid`, `name`, `event_type`) is modeled; a
  hypothetical record that only carries an alternate key name for one of these fields would see
  that field come through as `null` here where legacy would have populated it. Documented scope
  narrowing, not a silent divergence: no fixture or live RD Station response encountered during
  this migration exercised the alternate key names, and legacy's own fallback order always tries
  the primary name first.
- **`pagination.next_page` body signal is not read.** See Streams notes: the engine's
  `page_number` paginator stops on a short page only, not on an explicit empty/non-advancing
  `next_page` body value. Benign for any real RD Station response (a short page and an empty
  `next_page` co-occur in practice), but not proven identical for a hypothetical page that returns
  a full `page_size` count with an empty `next_page` value.
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size` (1-125,
  default 125) and `max_pages` (0/all/unlimited or a positive integer cap) as config-driven
  overrides (`pageSize`/`maxPages`, `rd_station_marketing.go:243-265`). The engine's `page_number`
  paginator has no config-driven request-count-cap knob at all, and since legacy's own
  `page_size` default already equals its own hard max (125), there is no effective config range
  lost by not wiring it — this bundle sends RD Station's own default/max (`page_size=125`)
  as the static pagination block.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (reached only
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance) stamps
  synthetic fields (`title`, `event_identifier`, etc.) that are not part of the LIVE record shape;
  this bundle's schemas and fixtures target the live path only (`harvest`), matching
  conventions.md's instruction to ignore legacy's fixture-mode-only fields.
