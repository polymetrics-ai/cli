# Overview

Flowlu is a CRM/project-management suite. This bundle reads six read-only streams — `accounts`,
`leads`, `tasks`, `projects`, `invoices`, `agile_issues` — from the Flowlu REST API v1
(`https://{company}.flowlu.com/api/v1/module/<resource>/list`). It migrates
`internal/connectors/flowlu` (the hand-written legacy connector), which stays registered and
unchanged until wave6's registry flip. Flowlu has no write/mutation surface exposed for reverse ETL
in this connector, matching legacy: `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Auth setup

Flowlu authenticates every request with an `api_key` query parameter (`streams.json` `base.auth`'s
`api_key_query` mode, `param: "api_key"`), sourced from the required `api_key` secret. The base URL
is derived from the required `company` config value (your Flowlu account subdomain) as
`https://{company}.flowlu.com/api/v1/module`, matching legacy's `flowluBaseURL` derivation
(`flowlu.go:273-296`) exactly — `streams.json` `base.url` templates `company` directly
(`https://{{ config.company }}.flowlu.com/api/v1/module`).

## Streams notes

All six streams share an identical shape: `GET <module>/<entity>/list`, records extracted from the
nested `response.items` path (`flowluRecordsPath` in legacy), primary key `id`. Pagination is
`page`/`count` page-number pagination (`page_number` type, `page_param: "page"`, `size_param:
"count"`), stopping on a short page (fewer than `page_size` records), matching legacy's `harvest`
loop (`flowlu.go:152-186`) exactly — legacy has no server-side incremental filter parameter, so
every read is a full stream read (matching legacy's `InitialState`, which always starts from an
empty cursor and never advances it via a request param).

Every schema declares `x-cursor-field: updated_date` for manifest-surface parity with legacy's
declared `CursorFields: []string{"updated_date"}` on every stream, but (matching legacy) no stream
declares an `incremental` block — neither this bundle nor legacy actually filters or advances reads
by that field; it is documentation of legacy's manifest shape only, not a functioning incremental
capability.

## Write actions & risks

None. Flowlu's write/mutation surface is not exposed by this connector (legacy: `Capabilities.Write:
false`, `Write` returns `ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle
ships no `writes.json`.

## Known limits

- **`base_url` override is not modeled.** Legacy accepts an optional `base_url` config override that
  bypasses the `company`-derived URL entirely (`flowluBaseURL`, `flowlu.go:273-296`: any absolute
  http/https URL with a host). The engine's `streams.json` `base.url` is a single fixed template with
  no conditional-override mechanism (unlike `auth`'s `when`-gated candidate list), so only the
  `company`-derivation path is modeled. `base_url` is not declared in `spec.json` at all (a
  declared-but-unwireable key is worse than an absent one — see `docs/migration/conventions.md`'s
  "declared config must be consumed" rule); operators who need a non-standard Flowlu host are out of
  scope for this bundle.
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy accepts optional `page_size`
  (1-100, default 100) and `max_pages` (default unlimited) config keys read at request time
  (`flowluPageSize`/`flowluMaxPages`, `flowlu.go:314-342`). The engine's `PaginationSpec.PageSize`/
  `MaxPages` fields are plain fixed JSON integers baked into `streams.json`'s `base.pagination` block
  — there is no templating/config-driven override mechanism for them at all. This bundle declares a
  fixed `page_size: 2` (chosen small so the required 2-page conformance fixture is realistic and
  exercises the short-page stop rule; legacy's own default is 100) and no `max_pages` cap
  (unbounded, matching legacy's own default). Neither `page_size` nor `max_pages` is declared in
  `spec.json` (F6: dead, unwireable config is worse than absent config).
