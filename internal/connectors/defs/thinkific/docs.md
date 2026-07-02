# Overview

Thinkific is a wave2 fan-out declarative-HTTP migration of `internal/connectors/thinkific` (the
hand-written legacy connector this bundle migrates; the legacy package stays registered and
unchanged until wave6's registry flip). It reads Thinkific courses through the Thinkific public API
(`GET https://api.thinkific.com/api/public/v1/courses`). Read-only.

## Auth setup

Provide a Thinkific API key via the `api_key` secret and the site subdomain via the `subdomain`
config value; both are sent as static request headers (`X-Auth-API-Key`/`X-Auth-Subdomain`,
`streams.json` `base.headers`), matching legacy's `connsdk.Requester.DefaultHeaders` construction
(`thinkific.go:110`: `{"X-Auth-API-Key": key, "X-Auth-Subdomain": subdomain}`). `base.auth`
declares `[{"mode": "none"}]` since Thinkific's auth is header-only, not a bearer/basic/api-key-*
mode the dialect's `auth` block otherwise expresses. `api_key` is never logged (`x-secret: true`).
`base_url` defaults to `https://api.thinkific.com` and may be overridden for tests/proxies.

## Streams notes

`courses` is the only stream: `GET /api/public/v1/courses`, records at `items`, primary key
`["id"]`. Every emitted field (`id`, `name`, `slug`, `created_at`) matches the raw API's own field
names exactly — no `computed_fields` rename needed, plain schema projection reproduces legacy's
inline record construction (`thinkific.go:86`) field-for-field. `id` is Thinkific's real wire type
(a JSON integer, not a string) — the schema declares `"type": "integer"`, matching the fixture's
real numeric shape, not a widened `["integer","string"]` union.

Pagination follows a 1-based page-number convention (`pagination.type: page_number`, `page_param:
page`, `size_param: limit`, `page_size: 100`) — matches legacy's hand-rolled loop
(`thinkific.go:72-93`: `page`/`limit` query params, stopping when a page returns fewer records than
the configured size).

## Write actions & risks

None. Legacy `thinkific` is read-only (`Write` returns `connectors.ErrUnsupportedOperation`);
`metadata.json` declares `capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- Full Thinkific API surface (users, enrollments, categories, products, orders) is out of scope
  for wave2; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries. Only the single legacy-parity `courses` stream is implemented.
- **`page_size` is not runtime-configurable.** Legacy exposes a config-driven `page_size` override
  (`thinkific.go:124-130`, `pageSize(cfg, 100)`, any positive integer, defaulting to 100 when unset
  or invalid). The engine's `page_number` paginator constructor reads `PaginationSpec.PageSize` as
  a static bundle-level integer from `streams.json`, not a config-templated field, so there is no
  mechanism to make it runtime-configurable from `config.page_size` without inventing Go. This
  bundle hardcodes `page_size: 100`, legacy's own default, matching every input that does not
  explicitly override the page size (the common case); an operator who previously set a
  smaller/larger `page_size` config value loses that override here. `page_size` is not declared in
  `spec.json` at all (F6, REVIEW.md: a declared-but-unwireable key is worse than an absent one).
- No incremental cursor is modeled. Legacy's catalog declares `created_at` as a `CursorFields`
  hint but the `Read` loop never actually filters by it (no incremental request param is ever
  sent) — this bundle matches that exact behavior: `schemas/courses.json` declares
  `x-cursor-field: created_at` for catalog-hint parity, but no `incremental` block is declared on
  the `courses` stream, so every sync is full-refresh, exactly like legacy.
