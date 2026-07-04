# Overview

The Guardian API is a wave2 fan-out declarative-HTTP migration of
`internal/connectors/the-guardian-api` (the hand-written legacy connector this bundle migrates;
the legacy package stays registered and unchanged until wave6's registry flip), since Pass-B
expanded to the full documented Guardian Open Platform Content API surface: content search, tags,
sections, editions, and single-item retrieval. Read-only — the Open Platform offers no write/
mutation capability at any tier.

## Auth setup

Provide a Guardian Open Platform API key via the `api_key` secret; it is sent as the `api-key`
query parameter (`api_key_query` auth mode), matching legacy's
`connsdk.APIKeyQuery("api-key", key)` (`the_guardian_api.go:116`). Never logged. `base_url`
defaults to `https://content.guardianapis.com` and may be overridden for tests/proxies.

## Streams notes

Five streams, matching the full documented Guardian Open Platform Content API
(conventions.md §"api_surface.json depth" Pass B expansion; see `api_surface.json`):

- `search` — `GET /search`, records at `response.results`, primary key `["id"]`. This is the
  original legacy-parity stream: the raw API's `id` field is already the legacy output field name
  (no rename needed); `webTitle` and `webPublicationDate` are renamed via `computed_fields` to
  `title`/`published_at`, matching legacy's `mapRecord`-equivalent inline construction
  (`the_guardian_api.go:93`). The optional free-text search query (`config.query`, legacy's `q`
  param, only sent when non-empty — `the_guardian_api.go:74-76`) is expressed via the opt-in
  optional-query object dialect (`"query": {"q": {"template": "{{ config.query }}",
  "omit_when_absent": true}}`, conventions.md §3): the `q` parameter is left off the request
  entirely when `query` is unset, exactly matching legacy's conditional branch, and sent verbatim
  when set. Pagination follows a 1-based page-number convention (`pagination.type: page_number`,
  `page_param: page`, `size_param: page-size`, `page_size: 50`) — matches legacy's hand-rolled loop
  (`the_guardian_api.go:77-100`), stopping when a page returns fewer records than the configured
  size.
- `tags` — `GET /tags` (documentation/md/tag.md), records at `response.results`, primary key
  `["id"]`. Shares the same `q` optional-query dialect and inherits the base `page_number`
  pagination (the docs' own example shows an identical `page`/`page-size` convention, defaulting to
  `page-size: 10` server-side when unset; this bundle always sends the base's `page_size: 50` like
  `search`, within the documented 1-50 accepted range — the docs do not state an upper bound for
  `/tags` specifically, so the same client-chosen page size used elsewhere in this bundle is
  reused). Fields: `id`/`type`/`webTitle`/`webUrl`/`apiUrl`/`sectionId`/`sectionName`, exactly the
  documented response fields.
- `sections` — `GET /sections` (documentation/md/section.md), records at `response.results`,
  primary key `["id"]`. The documented example response returns every section in one page with no
  page/page-size parameters listed for this endpoint at all; this stream declares
  `"pagination": {"type": "none"}` (stream-level override of the base `page_number` spec — a single
  request, no page/page-size query params sent) to match the documented shape honestly rather than
  sending an undocumented `page`/`page-size` pair the API may simply ignore. `editions` is returned
  nested per-section (an array of edition sub-objects); those are preserved verbatim inside each
  section's raw `editions` field rather than fanned out into separate records, since the top-level
  `/editions` endpoint already exposes the same edition objects as first-class records.
- `editions` — `GET /editions` (documentation/md/edition.md), records at `response.results`,
  primary key `["id"]`. Same `"pagination": {"type": "none"}` reasoning as `sections` — the
  documented example returns all editions (a small, fixed set: au/europe/international/uk/us) in
  one unpaginated response.
- `content` — `GET /{content-id}` (documentation/md/item.md, the "Single Item" endpoint), a
  single-object stream keyed by the new `config.content_id` spec property (the Guardian content
  path, e.g. `world/2024/jan/01/example-article` — there is no discovery/list endpoint for this
  path other than `search`'s own `id` field, so a caller supplies one explicitly, exactly like
  xkcd's `comic_number`-keyed `comic` stream in this same migration wave). `records: {"path":
  "response.content", "single_object": true}` selects the nested `content` object per the
  documented response envelope; `computed_fields` renames `webTitle`/`webPublicationDate` to
  `title`/`published_at` identically to `search`, keeping the two streams' emitted shapes
  consistent. `"pagination": {"type": "none"}` — a single-item lookup is never paginated.

## Write actions & risks

None. The Guardian Open Platform Content API is a public read-only content syndication API; it
offers no write/mutation endpoint at any access tier (confirmed against the full documented
surface, `api_surface.json`). `metadata.json` declares `capabilities.write: false` and this bundle
ships no `writes.json`.

## Known limits

- The content `/next` deep-pagination continuation endpoint
  (documentation/md/content_search.md's "Deep pagination" section) is not modeled as a separate
  stream: it returns the identical `search` record shape, addressed by the previous page's last
  result `id` rather than `page`/`page-size`, and exists purely as a workaround for exceeding
  `search`'s ordinary page-number pagination depth limit. `search`'s existing `page_number`
  pagination already covers the documented, supported page range; see `api_surface.json`'s
  `duplicate_of` exclusion.
- **`page_size` is not runtime-configurable.** Legacy exposes a config-driven `page_size` override
  (`the_guardian_api.go:130-136`, `pageSize(cfg, 50)`, any positive integer, defaulting to 50 when
  unset or invalid). The engine's `page_number` paginator constructor reads
  `PaginationSpec.PageSize` as a static bundle-level integer from `streams.json`, not a
  config-templated field, so there is no mechanism to make it runtime-configurable from
  `config.page_size` without inventing Go. This bundle hardcodes `page_size: 50`, legacy's own
  default, matching every input that does not explicitly override the page size (the common case);
  an operator who previously set a smaller/larger `page_size` config value loses that override
  here. `page_size` is not declared in `spec.json` at all (F6, REVIEW.md: a declared-but-unwireable
  key is worse than an absent one). This applies identically to the new `tags` stream (same base
  pagination block); `sections`/`editions`/`content` are unpaginated so this limit does not apply
  to them.
- No incremental cursor is modeled on any stream. Legacy's catalog declares `published_at` as a
  `CursorFields` hint but the `Read` loop never actually filters by it (no `from-date`/`to-date`
  request param is ever sent) — this bundle matches that exact behavior on `search`:
  `schemas/search.json` declares `x-cursor-field: published_at` for catalog-hint parity, but no
  `incremental` block is declared on the `search` stream, so every sync is full-refresh, exactly
  like legacy. The Content Search API does document `from-date`/`to-date`/`use-date` filter
  parameters that could drive a genuine server-side incremental filter in a future increment; this
  bundle does not wire them (out of scope for this pass, not an engine gap — `stream.Query`'s
  optional-query dialect could express it, this is a scope choice, not a blocker).
