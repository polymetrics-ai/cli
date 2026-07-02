# Overview

NewsData.io is a news search API exposing latest news, cryptocurrency news, and a source
directory. This bundle reads `/1/latest`, `/1/crypto`, and `/1/sources`, migrated from
`internal/connectors/newsdata` (the hand-written connector this bundle replaces at capability
parity); the legacy package stays registered and unchanged until wave6's registry flip. The
connector is read-only.

## Auth setup

Provide the `api_key` secret; the engine's declarative `api_key_query` auth mode appends
`apikey={{ secrets.api_key }}` to every request's query string, matching legacy's
`connsdk.APIKeyQuery("apikey", secret)`. The secret only ever flows into that query parameter and
is never logged.

Set `base_url` to override the API host; it defaults to `https://newsdata.io/api/1` (legacy's
`newsdataDefaultBaseURL`).

## Streams notes

Three streams, matching legacy's per-stream endpoint routing exactly:

- `latest` (`GET /latest`, records at `results`) — the latest-news feed. Primary key
  `article_id`, cursor field `pubDate` (recorded on the schema for catalog-metadata parity; see
  Known limits for why no `incremental` block is declared).
- `crypto` (`GET /crypto`, records at `results`) — same shape, cryptocurrency-scoped.
- `sources` (`GET /sources`, records at `results`) — the source directory; primary key `id`, no
  pagination (matching legacy's `endpoint.paginated == false` for this stream), no cursor field
  (matching legacy's own catalog, which declares no `CursorFields` for sources).

Every stream forwards the same optional filters via the optional-query dialect
(`omit_when_absent: true`, matching legacy's `newsdataFilters` which skips empty config values
entirely): `category`, `country`, `language`, `domain` (comma-joined restriction lists), `query`
(sent as `q`), `query_in_title` (sent as `qInTitle`), and `size`.

Pagination for `latest`/`crypto` is `cursor` with `token_path: nextPage` and `cursor_param: page`
(NewsData.io's own next-page-token convention: the response's `nextPage` value, when non-null, is
echoed back as the `page` query parameter on the following request) — matching legacy's `harvest`
loop exactly, including the stop condition (an absent/empty `nextPage` stops the loop, the
engine's default token_path behavior; no `stop_path` is needed since NewsData.io's own
`nextPage: null` is already representable by the token_path paginator's built-in
falsy-token-stops rule).

`max_pages` is capped at 5 (`pagination.max_pages: 5`), matching legacy's own default bound
(`newsdataDefaultMaxPages`) — see Known limits for why this bound is no longer
config-overridable.

## Write actions & risks

None. NewsData.io is read-only in both legacy and this bundle (`capabilities.write: false`); no
`writes.json` file is shipped. Legacy's own comment explains why: NewsData.io is a news search API
with no reverse-ETL surface.

## Known limits

- Legacy accepted a `max_pages` config override (`all`/`unlimited`/`0` for unbounded, or a
  positive integer, defaulting to 5 when unset). The engine's `PaginationSpec.MaxPages` is a fixed
  bundle-declared literal with no config-driven override mechanism in this dialect (matching
  nebius-ai's identical documented limitation) — this bundle narrows the config surface to the
  fixed default-5 cap only. This narrows accepted CONFIGURATION surface only, never emitted record
  data for the common (default) path.
- `/1/archive` (paid-tier historical search) is out of scope for this wave, matching legacy's own
  first-cut scope; see `api_surface.json`.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting for
  NewsData.io, so none is added here either.
