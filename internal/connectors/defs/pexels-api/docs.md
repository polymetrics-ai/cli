# Overview

Pexels API is a wave2 fan-out declarative-HTTP migration, expanded in Pass B to the full documented
Pexels API surface. It reads Pexels photo/video search and curated/popular results, plus featured
and personal collections and their media, through the Pexels REST API
(`GET https://api.pexels.com/v1/...`). The original 4 streams (`photos`, `curated_photos`,
`videos`, `popular_videos`) are engine-vs-legacy parity-tested against
`internal/connectors/pexels-api` (the hand-written connector this bundle migrates); the legacy
package stays registered and unchanged until wave6's registry flip. The 3 Pass B streams
(`featured_collections`, `my_collections`, `collection_media`) have no legacy counterpart — legacy
never modeled Pexels Collections at all — so they carry no parity constraint and are authored
directly from Pexels' own published documentation.

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

**Pass B additions** (no legacy counterpart; authored from Pexels' own docs,
`https://www.pexels.com/api/documentation/#collections`):

- `featured_collections` (`GET /v1/collections/featured`, records at `collections`): Pexels'
  curated public collections. Same `next_url` pagination shape as every other stream.
- `my_collections` (`GET /v1/collections`, records at `collections`): the authenticated account's
  own collections (Pexels' docs: "This endpoint returns all of your collections").
- `collection_media` (`GET /v1/collections/{id}`, records at `media`): the photos/videos inside one
  collection. Pexels' docs state plainly that collections themselves cannot be created or listed by
  arbitrary id — there is no way to discover a collection id except by first listing collections —
  so this stream uses `fan_out` (`ids_from.request` against `my_collections`' own `GET
  /v1/collections` endpoint, `into.path_var: id`, `stamp_field: collection_id`) to drive
  one `/v1/collections/{id}` sub-sequence per collection this account owns, stamping the source
  collection's id onto every emitted media record. Each item in `media` carries Pexels' own `type`
  discriminator (`"Photo"` or `"Video"`) and the two shapes' fields differ (a `Photo` has `src`/
  `alt`; a `Video` has `video_files`/`video_pictures`/`duration`) — `collection_media` therefore
  declares `"projection": "passthrough"` (not schema-mode) so neither variant's fields are silently
  dropped; `schemas/collection_media.json` documents the union of both shapes as a reference
  surface only. Optional `collection_media_type` (`photos`/`videos`) and `collection_media_sort`
  (`asc`/`desc`) config values forward Pexels' own `type`/`sort` query filters when set, omitted
  otherwise (`omit_when_absent: true`).
- The lone by-id detail lookups `GET /v1/photos/{id}` and `GET /v1/videos/videos/{id}` are excluded
  (`requires_elevated_scope` in `api_surface.json`): unlike `collection_media`'s id (discoverable
  via `my_collections`), there is no Pexels endpoint that enumerates arbitrary photo/video ids
  independent of a search query the `photos`/`videos` streams already cover, so there is no
  server-side id source to drive either as a syncable stream.

## Write actions & risks

None. Legacy `pexelsapi.Write` always returns `connectors.ErrUnsupportedOperation`;
`capabilities.write` is `false` and this bundle ships no `writes.json`. This is not merely a
legacy-scope gap: Pexels' own documentation states plainly that "Collections cannot be created or
modified using the Pexels API," and no POST/PUT/PATCH/DELETE endpoint exists anywhere in the
documented Pexels API surface (Photos, Videos, or Collections) — the API is entirely read-only by
design, confirmed by direct inspection of the full published API reference during this Pass B
review.

## Known limits

- **`max_pages` is not runtime-configurable**, for the same reason as bitly/persona: the engine's
  `next_url` paginator has no config-driven page-count knob. `spec.json` does not declare
  `max_pages` (F6, REVIEW.md). Pagination is bounded only by the short/empty `next_page` stop
  signal, matching legacy's own real termination behavior. This applies identically to the 3 Pass B
  streams.
- **`page_size` is not runtime-configurable** because this bundle sends legacy's fixed
  `per_page=40` static query on first requests; declaring a `page_size` spec key would be dead
  config.
- **Fixtures are single-page** for every stream, per `docs/migration/conventions.md` §4's
  sanctioned `next_url` exception: Pexels' `next_page` URL is the replay server's own runtime
  address and cannot be embedded in a static fixture file. Every fixture omits `next_page`
  entirely, so `pagination_terminates` (which runs against the bundle's first eligible stream,
  `photos`) correctly observes exactly one request for one fixture page and terminates. Real
  2-page `next_page` correctness is proven by legacy's own `pexels_api_test.go`'s
  `TestReadPhotosPaginatesAndAuthenticates` (a live `httptest.Server` asserting the second page is
  requested via the exact absolute `next_page` URL legacy's response served); `collection_media`'s
  fan-out sub-sequence pagination is exercised the same way the fixture replay proves it (one page
  per collection id, no `next_page`), consistent with every other stream in this bundle.
- Full Pexels API surface is now covered: 7 of the 9 documented endpoints are syncable streams; the
  remaining 2 (`GET /v1/photos/{id}`, `GET /v1/videos/videos/{id}`) are excluded as
  `requires_elevated_scope` (no server-side id-discovery path independent of an already-covered
  search stream) — see `api_surface.json`.
