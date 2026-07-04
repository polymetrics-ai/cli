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
computed from the raw `articleId` field with legacy's `id` fallback, and `pub_date` is computed
from `pubDate` with legacy's `publishedAt` fallback. `title`, `url`, and `source` pass through
schema projection unchanged. Pagination is `page_number` (`page`/`size` query
params, matching legacy's `PageNumberPaginator{PageParam:"page", SizeParam:"size"}`), 100 records
per page by default. An optional `query` config value is sent as the `q` parameter when set
(legacy: `perigon.go:90-92`). An optional incremental lower bound (state cursor or `start_date`
config) is sent as the `from` parameter when it resolves (legacy: `firstNonEmpty(req.State["cursor"],
req.Config.Config["start_date"])` at `perigon.go:93-95`); no client-side filter is applied, matching
legacy's behavior of sending the parameter and emitting whatever the server returns.

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
