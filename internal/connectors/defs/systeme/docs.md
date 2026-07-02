# Overview

Systeme is a wave2 fan-out declarative-HTTP migration of `internal/connectors/systeme` (the
hand-written legacy connector this bundle migrates; the legacy package stays registered and
unchanged until wave6's registry flip). It reads Systeme.io contacts through the Systeme.io API
(`GET https://api.systeme.io/api/contacts`). Read-only.

## Auth setup

Provide a Systeme.io API key via the `api_key` secret; it is sent as the `X-API-Key` header
(`auth.mode: api_key_header`), matching legacy's `connsdk.APIKeyHeader("X-API-Key", secret, "")`
(`systeme.go:113`) — no prefix. Never logged. `base_url` defaults to
`https://api.systeme.io/api` and may be overridden for tests/proxies.

## Streams notes

`contacts` is the only stream: `GET /contacts`, records at `items`, primary key `["id"]`, cursor
field `created_at`. Systeme.io's real wire shape returns the creation timestamp as camelCase
`createdAt` (legacy's own `first()` helper prefers `created_at` and falls back to `createdAt`, but
the live API only ever emits `createdAt`); a `computed_fields` rename (`"created_at": "{{
record.createdAt }}"`) reproduces legacy's effective output field name exactly.

Pagination follows a 1-based page-number convention (`pagination.type: page_number`,
`page_param: page`, `size_param: limit`, `page_size: 100`) — matches legacy's
`connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "limit", StartPage: 1, PageSize:
pageSize}`, stopping when a page returns fewer records than `page_size` (a short page).

## Write actions & risks

None. Legacy `systeme` is read-only (`Write` returns `connectors.ErrUnsupportedOperation`);
`metadata.json` declares `capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- Full Systeme.io API surface (tags, courses, funnels, communities, email campaigns) is out of
  scope for wave2; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries. Only the single legacy-parity `contacts` stream is implemented.
- **`page_size` is not runtime-configurable.** Legacy exposes a config-driven `page_size` override
  (1-100, default 100; `systeme.go:182-192`). The engine's `page_number` paginator constructor
  reads `PaginationSpec.PageSize` as a static bundle-level integer, resolved once at read-start
  from a hardcoded fallback constant when unset (`read.go`'s `readDeclarative`), never from a
  `config.*` reference — there is no mechanism to template it. This bundle bakes in
  `page_size: 100`, legacy's own default, matching every input that does not explicitly override
  the page size (the common case); an operator who previously set a smaller/larger `page_size`
  config value loses that override here. `page_size` is not declared in `spec.json` at all (F6,
  REVIEW.md: a declared-but-unwireable config key is worse than an absent one).
- All fixtures (`fixtures/streams/contacts/**`, `fixtures/check.json`) represent Systeme.io's real
  wire shape, including the camelCase `createdAt` field.
