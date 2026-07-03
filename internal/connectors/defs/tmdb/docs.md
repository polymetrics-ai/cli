# Overview

TMDb is a wave2 fan-out declarative-HTTP migration. It reads popular movies, now-playing movies,
movie search results, and single-movie details from The Movie Database API
(`GET https://api.themoviedb.org/3/...`). This bundle targets capability parity with
`internal/connectors/tmdb` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a TMDb v3 API key via the `api_key` secret; it is sent as the `api_key` query parameter
(`api_key_query` auth mode), matching legacy's `baseQuery`'s `url.Values{"api_key": {key}}`
(`tmdb.go:197-207`), and is never logged. An optional `language` config value is applied as the
`language` query parameter on every request when set (legacy's same conditional
`if lang := ...; lang != ""` behavior), omitted entirely otherwise.

## Streams notes

`popular_movies` and `now_playing_movies` share the identical TMDb list envelope
(`GET /movie/popular` / `GET /movie/now_playing` returning `{"results":[...],"page":N,
"total_pages":N,...}`); pagination is `page_number` (`page` param only, no page-size query
parameter — TMDb's list size is a fixed 20 server-side and legacy's own `PageNumberPaginator`
construction never set a `SizeParam`, `tmdb.go:123`) starting at page 1, static `page_size: 20`
matching legacy's `defaultPageSize` short-page stop threshold.

`search_movies` (`GET /search/movie`) shares the same envelope and pagination shape but additionally
requires the `query` config value, sent as the `query` search-text query parameter — legacy
requires this and errors if unset (`tmdb.go:93-97`); this bundle's `query` param is a plain
(non-optional) template, so an unset `query` config hard-errors identically at read time.

`movie_details` (`GET /movie/{movie_id}`) is TMDb's only non-paginated, non-list stream: it returns
a single movie object directly at the response root, not wrapped in a `results` envelope. This
bundle declares `records.path: ""` (root-as-single-record, matching legacy's `recordsPath: "."`)
and a stream-level `pagination: {"type": "none"}` override (legacy's `!spec.paginated` branch,
`tmdb.go:99-113`, issues exactly one request with no page loop at all). `movie_id` is a required
path-templated config value scoped to this one stream only (not globally required in `spec.json`,
since the other three streams never reference it) — an unset `movie_id` hard-errors at read time
exactly like legacy's `movieDetailPath` (`tmdb.go:178-184`).

None of the four streams declare an `incremental` block: legacy's `Read` never applies a
cursor-based filter parameter of any kind — every read is either a full paginated sweep or (for
`movie_details`) a single fixed-resource fetch, matching legacy's true behavior exactly.

Every stream uses `projection: "passthrough"`: legacy's `Read` emits each record verbatim via
`emit(connectors.Record(rec))` for both the paginated-list branch (`tmdb.go:109`) and the
`Harvest`-driven branch (`tmdb.go:124`), with no field-building step in either case. Schema-mode
projection would silently drop any TMDb response field not in the declared list; the schemas here
document legacy's own known field surface (legacy's `fields(...)` declarations, `tmdb.go:153-156`)
but do not constrain what is actually emitted.

## Write actions & risks

None. TMDb is read-only (`capabilities.write` is `false`); this bundle ships no `writes.json`,
matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size` (default
  20, positive integer) and `max_pages` (0/absent = unbounded, non-negative integer cap) as
  config-driven overrides (`pageSize`/`maxPages`, `tmdb.go:225-246`), applied to the three paginated
  streams. The engine's `page_number` paginator has no config-driven page-size or
  request-count-cap knob (mirrors the aha/thinkific-courses/ticketmaster precedent from this same
  wave); `page_size`/`max_pages` are therefore not declared in `spec.json`, and this bundle sends
  TMDb's own default page size (20, matched via the short-page stop threshold) with no page cap.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path stamps a
  `fixture: true` marker field with no live-path equivalent (`tmdb.go:190`). This bundle's schemas
  and fixtures target the live record shape only; the engine's own `internal/connectors/conformance`
  fixture-replay harness provides the credential-free test affordance this bundle needs.
