# Overview

Perigon is a wave2 fan-out declarative-HTTP migration, expanded to full documented API-surface
coverage in Pass B. It reads all 7 of Perigon's documented v1 search/list endpoints: news articles
and story clusters (the original legacy-parity streams), plus journalists, sources, companies,
people, and topics — Perigon's full public entity-search surface (`GET
https://api.perigon.io/v1/{articles,stories,journalists,sources,companies,people,topics}/all`).
Perigon publishes no write/mutation endpoints at all; it is a read-only news-intelligence API.
`internal/connectors/perigon` (the hand-written legacy connector this bundle migrates) only ever
implemented `articles`/`stories`; the 5 additional streams are new Pass-B capability, not a legacy
port — legacy stays registered and unchanged until wave6's registry flip.

## API surface (Pass B)

Every documented Perigon v1 endpoint is now `covered_by` a stream in `api_surface.json` — there are
no `excluded` entries left. Perigon's documented surface is exactly these 7 `/all` search
endpoints; there are no per-ID single-resource GETs, no webhook/callback registration endpoints,
and no POST/PUT/DELETE mutation endpoints documented anywhere in
`https://docs.goperigon.com/` (confirmed via the Products/Reference navigation: Articles, Stories,
People, Companies, Topics, Journalists, Sources — all `/all` search-style GETs only).

## Auth setup

Provide a Perigon API key via the `api_key` secret; it is sent as the `apiKey` query parameter
(`api_key_query` auth mode), matching legacy's `connsdk.APIKeyQuery("apiKey", key)`
(`perigon.go:151`). `base_url` defaults to `https://api.perigon.io` and may be overridden for
tests/proxies.

## Streams notes

`articles` (`GET /v1/articles/all`) maps the raw response's `articles[]` array. `article_id` is
computed from the raw `articleId` field and `pub_date` from the raw `pubDate` field via
`computed_fields`, matching legacy's `articleRecord` primary paths. `title`, `url`, and `source`
pass through schema projection unchanged. Pagination is `page_number` (`page`/`size` query
params, matching legacy's `PageNumberPaginator{PageParam:"page", SizeParam:"size"}`), 100 records
per page by default. An optional `query` config value is sent as the `q` parameter when set
(legacy: `perigon.go:90-92`). An optional incremental lower bound (state cursor or `start_date`
config) is sent as the `from` parameter when it resolves (legacy: `firstNonEmpty(req.State["cursor"],
req.Config.Config["start_date"])` at `perigon.go:93-95`); `client_filtered: true` since Perigon's
`from` parameter is not proven to guarantee strict server-side ordering by publish date across
pages, matching legacy's behavior of never re-validating server-side filtering beyond sending the
parameter.

`stories` (`GET /v1/stories/all`) is a passthrough stream: legacy's `passthroughRecord` emits the
raw item unchanged, so this bundle declares `"projection": "passthrough"` with no
`computed_fields`, and the stories schema documents Perigon's real wire field names
(`id`/`name`/`createdAt`/`updatedAt`) rather than legacy's stale `commonFields()` stream-catalog
declaration (which lists `created_at`/`updated_at` but is never actually applied by
`passthroughRecord` — the schema here follows what legacy's `mapRecord` function actually emits,
per `docs/migration/conventions.md`'s schema-as-projection rule).

`journalists`, `sources`, `companies`, and `people` (`GET /v1/journalists/all`,
`/v1/sources/all`, `/v1/companies/all`, `/v1/people/all`) are new Pass-B streams, each a
`passthrough`-projected `page_number` list read with records at `results`, matching the real
Perigon wire envelope (`{"status", "numResults", "results": [...]}`) for every one of these 4
endpoints. Each accepts the same optional `q` search-query parameter as `articles` (sent from the
shared `config.query` value, `omit_when_absent`); none of the 4 exposes any incremental/cursor
filter in Perigon's documented query-parameter set, so no `incremental` block is declared (parity
with the read-only, always-full-search nature of these entity-search endpoints — there is no
`updatedAt`-range filter parameter documented for any of them, only free-text/faceted search
params). `people`'s primary key is `wikidataId` (Perigon's own documented identifier for this
endpoint, not `id`); the other 3 use `id`.

`topics` (`GET /v1/topics/all`) is a `passthrough` stream with records at `data` (a small,
enumerable taxonomy list — Perigon's own docs describe it as "all the available topics... on all
currently available Articles" rather than a growing per-record collection); primary key is `name`
(topics have no numeric/opaque `id` in Perigon's documented shape). It carries no query parameters
beyond the shared base pagination, since Perigon's topics endpoint documents no search/filter
parameters of its own.

## Write actions & risks

None. Perigon publishes no write/mutation endpoint anywhere in its documented API surface (see API
surface note above) — this is not a narrowed scope, it is the entirety of what Perigon exposes.
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`article_id`/`pub_date` fallback paths are not modeled.** Legacy's `articleRecord` uses
  `firstAny(item, "articleId", "id")` and `firstAny(item, "pubDate", "publishedAt")` — a
  first-non-nil-of-two-paths fallback. The engine's `computed_fields` dialect has no
  multi-path-fallback primitive (only a single bare `{{ record.<path> }}` reference, a filter
  chain, or a static literal), so this bundle wires only the primary path (`articleId`/`pubDate`)
  of each pair. This is capability parity for every input Perigon's real API emits (its documented
  article shape always includes `articleId` and `pubDate`) and for every input legacy's own test
  suite (`perigon_test.go`) exercises — neither ever supplies the fallback-only shape. A
  hypothetical malformed record missing `articleId` entirely would silently drop `article_id` here
  (the field is skipped for that record, per the engine's absent-computed-field-source rule) rather
  than falling back to `id` as legacy would; this is an accepted, documented, non-reproducible-in-
  practice deviation, not a data-shape change for any real Perigon response.
- **`journalists`/`sources`/`companies`/`people`/`topics` have no incremental cursor.** Perigon's
  documented query parameters for these 5 endpoints include no `updatedAt`/date-range filter (only
  `articles` documents a publish-date-oriented `from` filter), so every read of these streams is a
  full re-fetch; `x-cursor-field` is intentionally absent from their schemas rather than declared
  as unused catalog metadata. This matches the real documented API, not a narrowing choice.
- **Rich nested response fields (`topSources`, `location`, `webResources`, `occupation`, etc.) are
  passed through verbatim, untyped beyond `array`/`object`.** These streams use
  `"projection": "passthrough"`, so every raw field Perigon returns survives; the schema's
  `array`/`object` typing for nested structures is intentionally loose (matching the engine's
  schema-as-documentation-only stance under passthrough projection) rather than fully modeling
  Perigon's nested sub-object shapes field-by-field.
