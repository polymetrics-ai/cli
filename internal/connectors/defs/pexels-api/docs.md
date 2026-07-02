# Overview

Pexels API is a wave2 fan-out declarative-HTTP migration. It reads Pexels photo and video search
results through the Pexels REST API (`GET https://api.pexels.com/v1/...`). This bundle is
engine-vs-legacy parity-tested against `internal/connectors/pexels-api` (the hand-written
connector it migrates); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide a Pexels API key via the `api_key` secret; it is sent as the RAW `Authorization` header
value with no `Bearer`/other prefix (`auth.mode: api_key_header`, `header: Authorization`, no
`prefix`) and is never logged, matching legacy's `connsdk.APIKeyHeader("Authorization", key, "")`
(`pexels_api.go:167`). `base_url` defaults to `https://api.pexels.com` and may be overridden for
tests/proxies.

## Streams notes

Four streams: `photos` (`GET /v1/search`, records at `photos`), `curated_photos` (`GET
/v1/curated`, records at `photos`), `videos` (`GET /v1/videos/search`, records at `videos`),
`popular_videos` (`GET /v1/videos/popular`, records at `videos`) — matching legacy's
`streamEndpoints` map exactly. `photos`/`videos` require a search `query`; legacy falls back to the
literal `"people"` when the config value is unset (`pexels_api.go:90-92`), expressed here as
`query`'s `default: "people"` on the object-form query param (per conventions.md §3's optional-
query dialect) rather than a `spec.json`-level default (the fallback is stream-specific: only
`photos`/`videos` apply it, not `curated_photos`/`popular_videos`, which legacy never defaults).
Every stream additionally forwards `orientation`/`size`/`color`/`locale` from config when set,
omitted entirely otherwise (`omit_when_absent: true`) — matching legacy's `for _, key :=
range [...]{ if v := ...; v != "" { query.Set(key, v) } }` loop (`pexels_api.go:85-89`), applied
identically across all 4 streams (legacy applies this loop unconditionally regardless of
`requiresQuery`).

Pagination follows Pexels' own `next_page` absolute-URL convention (`pagination.type: next_url`,
`next_url_path: "next_page"`), matching legacy's manual loop that follows `resp.Body`'s `next_page`
field verbatim until empty (`pexels_api.go:108-115`). `page=1`/`per_page=40` (legacy's own default
page size) are sent as static per-stream query values on the first request, matching stripe's
static-query precedent for the same reason the `next_url` paginator has no page-size knob
(conventions.md §3's pagination table).

## Write actions & risks

None. Legacy `pexelsapi.Write` always returns `connectors.ErrUnsupportedOperation`;
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`max_pages` is not runtime-configurable**, for the same reason as bitly/persona: the engine's
  `next_url` paginator has no config-driven page-count knob. `spec.json` does not declare
  `max_pages` (F6, REVIEW.md). Pagination is bounded only by the short/empty `next_page` stop
  signal, matching legacy's own real termination behavior.
- **Fixtures are single-page** for every stream, per `docs/migration/conventions.md` §4's
  sanctioned `next_url` exception: Pexels' `next_page` URL is the replay server's own runtime
  address and cannot be embedded in a static fixture file. Every fixture omits `next_page`
  entirely, so `pagination_terminates` (which runs against the bundle's first eligible stream,
  `photos`) correctly observes exactly one request for one fixture page and terminates. Real
  2-page `next_page` correctness is proven by legacy's own `pexels_api_test.go`'s
  `TestReadPhotosPaginatesAndAuthenticates` (a live `httptest.Server` asserting the second page is
  requested via the exact absolute `next_page` URL legacy's response served).
- Full Pexels API surface (single-photo lookup, featured collections) is out of scope; see
  `api_surface.json`'s `excluded: {category: out_of_scope}` entries — legacy itself never
  implemented these.
