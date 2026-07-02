# Overview

SimFin is a wave2 fan-out declarative-HTTP migration. It reads SimFin companies, statements, and
markets through the SimFin REST API (`GET https://backend.simfin.com/api/v3/...`). This bundle
targets capability parity with `internal/connectors/simfin` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a SimFin API key via the `api_key` secret; it is sent as the `api-key` query parameter on
every request (`mode: api_key_query`) and is never logged, matching legacy's
`connsdk.APIKeyQuery("api-key", token)` (`simfin.go:123`). `base_url` defaults to
`https://backend.simfin.com` and may be overridden for tests/proxies.

## Streams notes

All 3 streams (`companies`, `statements`, `markets`) are `GET` list endpoints
(`api/v3/companies/list`, `api/v3/statements/list`, `api/v3/markets/list`) with records at the
top-level `data` key, primary key `["id"]`. None declares an incremental cursor, matching legacy's
`streams()` exactly (no `CursorFields` set for any of the three). Pagination follows a 1-based
page-number convention (`pagination.type: page_number`, `page_param: page`, `size_param: limit`,
`start_page: 1`, `page_size: 100`) with the engine's standard short-page stop (a page returning
fewer than 100 records ends the read) — identical to legacy's own
`len(records) < pageSize` stop condition in `harvest`.

`statements` records carry a computed `updated_at` field sourced from the raw `fiscalPeriod` value
(a bare `{{ record.fiscalPeriod }}` reference, typed-extraction passthrough) — legacy's shared
`simfinRecord` mapper's `first(item, "updated_at", "fiscalPeriod")` fallback, narrowed to the single
candidate SimFin's statements endpoint actually emits (statements are versioned by fiscal period,
never carry a genuine `updated_at` field on the real wire shape). `companies`/`markets` declare no
`updated_at` computed field: their real wire shape carries neither `updated_at` nor `fiscalPeriod`
(company/market metadata, not a financial statement), so schema projection alone (a no-op passthrough
that leaves `updated_at` absent) matches legacy's `first()` call resolving to `nil` on both sides.

## Write actions & risks

None. SimFin's read endpoints have no obviously-safe reverse-ETL writes modeled; `capabilities.write`
is `false` and this bundle ships no `writes.json`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`pageSize`/`maxPages` helpers, `simfin.go:204-228`). The engine's `page_number`
  paginator reads `PaginationSpec.PageSize`/`MaxPages` only from the bundle's own static
  `streams.json` (never templated from `config.*`), so a `spec.json`-declared `page_size`/`max_pages`
  property would be genuinely dead config (F6, REVIEW.md — a declared-but-unwireable key is worse
  than an absent one). This bundle hardcodes legacy's own compiled-in default (`page_size: 100`,
  matching `simfinDefaultPageSize`) as a static `pagination.page_size` literal instead, and does not
  declare `page_size`/`max_pages` in `spec.json` at all. Pagination is unbounded (no `max_pages` cap),
  matching legacy's own default (`max_pages` unset -> unlimited).
- **Cross-endpoint `updated_at` fallback is narrowed to `statements` only.** See "Streams notes"
  above — this is a documented, verified-equivalent expression of legacy's shared `simfinRecord`
  mapper for every record shape SimFin's real API actually emits, not an arbitrary scope narrowing.
