# Overview

The Guardian API is a wave2 fan-out declarative-HTTP migration of
`internal/connectors/the-guardian-api` (the hand-written legacy connector this bundle migrates;
the legacy package stays registered and unchanged until wave6's registry flip). It reads Guardian
content search results through the Guardian Open Platform Content API
(`GET https://content.guardianapis.com/search`). Read-only.

## Auth setup

Provide a Guardian Open Platform API key via the `api_key` secret; it is sent as the `api-key`
query parameter (`api_key_query` auth mode), matching legacy's
`connsdk.APIKeyQuery("api-key", key)` (`the_guardian_api.go:116`). Never logged. `base_url`
defaults to `https://content.guardianapis.com` and may be overridden for tests/proxies.

## Streams notes

`search` is the only stream: `GET /search`, records at `response.results`, primary key `["id"]`.
The raw API's `id` field is already the legacy output field name (no rename needed); `webTitle` and
`webPublicationDate` are renamed via `computed_fields` to `title`/`published_at`, matching legacy's
`mapRecord`-equivalent inline construction (`the_guardian_api.go:93`).

The optional free-text search query (`config.query`, legacy's `q` param, only sent when
non-empty — `the_guardian_api.go:74-76`) is expressed via the opt-in optional-query object
dialect (`"query": {"q": {"template": "{{ config.query }}", "omit_when_absent": true}}`,
conventions.md §3): the `q` parameter is left off the request entirely when `query` is unset,
exactly matching legacy's `if query := ...; query != "" { base.Set("q", query) }` branch, and sent
verbatim when set.

Pagination follows a 1-based page-number convention (`pagination.type: page_number`, `page_param:
page`, `size_param: page-size`, `page_size: 50`) — matches legacy's hand-rolled loop
(`the_guardian_api.go:77-100`: `page`/`page-size` query params, stopping when a page returns fewer
records than the configured size).

## Write actions & risks

None. Legacy `the-guardian-api` is read-only (`Write` returns
`connectors.ErrUnsupportedOperation`); `metadata.json` declares `capabilities.write: false` and
this bundle ships no `writes.json`.

## Known limits

- Full Guardian Content API surface (single-item retrieval, tags, sections, editions) is out of
  scope for wave2; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries. Only the single legacy-parity `search` stream is implemented.
- **`page_size` is not runtime-configurable.** Legacy exposes a config-driven `page_size` override
  (`the_guardian_api.go:130-136`, `pageSize(cfg, 50)`, any positive integer, defaulting to 50 when
  unset or invalid). The engine's `page_number` paginator constructor reads
  `PaginationSpec.PageSize` as a static bundle-level integer from `streams.json`, not a
  config-templated field, so there is no mechanism to make it runtime-configurable from
  `config.page_size` without inventing Go. This bundle hardcodes `page_size: 50`, legacy's own
  default, matching every input that does not explicitly override the page size (the common case);
  an operator who previously set a smaller/larger `page_size` config value loses that override
  here. `page_size` is not declared in `spec.json` at all (F6, REVIEW.md: a declared-but-unwireable
  key is worse than an absent one).
- No incremental cursor is modeled. Legacy's catalog declares `published_at` as a `CursorFields`
  hint but the `Read` loop never actually filters by it (no `from-date`/`to-date` request param is
  ever sent) — this bundle matches that exact behavior: `schemas/search.json` declares
  `x-cursor-field: published_at` for catalog-hint parity, but no `incremental` block is declared on
  the `search` stream, so every sync is full-refresh, exactly like legacy.
