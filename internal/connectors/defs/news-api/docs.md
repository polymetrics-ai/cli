# Overview

News API (newsapi.org) is a news-article and source search API. This bundle reads the
`/v2/everything` search, `/v2/top-headlines`, and the `/v2/top-headlines/sources` directory,
migrated from `internal/connectors/news-api` (the hand-written connector this bundle replaces at
capability parity); the legacy package stays registered and unchanged until wave6's registry flip.
The connector is read-only.

## Auth setup

Provide the `api_key` secret; the engine's declarative `api_key_header` auth mode sends
`X-Api-Key: {{ secrets.api_key }}` on every request, matching legacy's
`connsdk.APIKeyHeader("X-Api-Key", key, "")`. The secret only ever flows into that header and is
never logged.

Set `base_url` to override the API host; it defaults to `https://newsapi.org/v2` (legacy's
`defaultBaseURL`).

## Streams notes

Three streams, matching legacy's per-stream endpoint/filter routing exactly:

- `everything` (`GET /everything`, records at `articles`) — the free-text article search. Every
  legacy-forwarded filter (`q`, `searchIn`, `sources`, `domains`, `excludeDomains`, `language`,
  `sortBy`, `to`) is wired via the optional-query dialect (`omit_when_absent: true`), sent only
  when its corresponding config value is set, matching legacy's `set()` helper which skips empty
  values entirely. `from` is wired through the `incremental` block instead
  (`request_param: from`, `start_config_key: start_date`, `cursor_field: published_at`): it
  resolves to the state cursor when a repeat sync has one, falling back to the `start_date` config
  value on a fresh sync, and is omitted entirely when neither resolves — matching legacy's
  `queryParams` (which sets `from` from `start_date` and then overrides it with the incremental
  cursor when present) exactly.
- `top_headlines` (`GET /top-headlines`, records at `articles`) — same article shape, filtered by
  `q`/`country`/`category`/`sources`/`language` (all optional-query); no incremental cursor is
  wired (legacy's own catalog gives it `CursorFields: ["published_at"]` but `queryParams` never
  applies any date filter to this endpoint — matching that, no `incremental` block is declared here
  either, though the schema still records `x-cursor-field: published_at` for catalog-metadata
  parity, per dwolla's identical precedent).
- `sources` (`GET /top-headlines/sources`, records at `sources`) — the source directory, filtered
  by `category`/`language`/`country` (all optional-query); no pagination, no incremental cursor
  (matching legacy's own catalog, which declares no `CursorFields` for this stream).

Both article streams hoist the nested `{source:{id,name}}` object to `source_id`/`source_name` and
rename the API's camelCase `urlToImage`/`publishedAt` to `url_to_image`/`published_at` via
`computed_fields`, matching legacy's `articleRecord` mapper exactly. Pagination for both article
streams is `page_number` (`page`/`pageSize`, `start_page: 1`, `page_size: 100`), matching legacy's
`connsdk.PageNumberPaginator` configuration verbatim (short-page stop).

## Write actions & risks

None. News API is read-only in both legacy and this bundle (`capabilities.write: false`); no
`writes.json` file is shipped.

## Known limits

- Legacy accepted a `page_size` config override (1-100, default 100) for the paginated streams'
  `pageSize` query param. `PaginationSpec.PageSize` is a fixed bundle-declared literal with no
  config-driven override mechanism in this dialect (matching stripe's own `limit=100` static-query
  precedent — see `docs/migration/conventions.md`'s ledger item 3) — this bundle fixes `pageSize`
  at legacy's own default of 100 with no `spec.json` override property. This narrows accepted
  CONFIGURATION surface only, never emitted record data for the common (default) path.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting for
  News API, so none is added here either.
- Full News API surface (this bundle implements exactly legacy's 3 streams; News API itself
  exposes no other read endpoints) — see `api_surface.json`.
