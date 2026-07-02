# Overview

Perigon is a wave2 fan-out declarative-HTTP migration. It reads Perigon news articles and story
clusters through the Perigon REST API (`GET https://api.perigon.io/v1/articles/all` and
`/v1/stories/all`). This bundle is a capability-parity port of `internal/connectors/perigon` (the
hand-written connector it migrates); the legacy package stays registered and unchanged until
wave6's registry flip.

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

## Write actions & risks

None. Perigon's legacy connector is read-only (`Capabilities.Write: false`); this bundle ships no
`writes.json`.

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
