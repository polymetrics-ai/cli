# Overview

Instatus is a wave2 fan-out declarative-HTTP migration. It reads Instatus status pages,
components, incidents, and maintenances through the Instatus REST API
(`https://api.instatus.com`). This bundle is engine-vs-legacy parity-tested against
`internal/connectors/instatus` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide an Instatus API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's
`connsdk.Bearer(secret)`. `base_url` defaults to `https://api.instatus.com` and may be overridden
for tests/proxies.

## Streams notes

`pages` is a top-level list endpoint (`GET /v2/pages`); records are a top-level JSON array
(`records.path: ""`). `components` (`GET /v2/{page_id}/components`), `incidents` (`GET
/v1/{page_id}/incidents`), and `maintenances` (`GET /v2/{page_id}/maintenances`) are parent-scoped:
the required `page_id` config value is substituted into the path (urlencoded by
`InterpolatePath`'s per-segment default, matching legacy's own `url.PathEscape(pageID)` in
`instatusPath`); an absent `page_id` hard-errors on both sides (legacy: `"instatus stream %q
requires config page_id"`; engine: an unresolved `config.page_id` path-template key — same failure
classification, different literal text). All four streams share the base `page`/`per_page`
pagination (`pagination.type: page_number`, `page_param: page`, `size_param: per_page`,
`start_page: 1`), stopping when a page returns fewer than `page_size` records — legacy's exact
`len(records) < pageSize` short-page stop rule. None of the four Instatus resources expose an
incremental cursor field (the legacy package's own doc: "the API only supports full-refresh
syncs"), so no stream declares an `incremental` block, matching legacy exactly.

## Write actions & risks

None. Instatus is a status-page monitoring API with no reverse-ETL surface here;
`capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's `Write`
returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`page_size` is not runtime-configurable in the same way legacy allowed.** Legacy exposes
  `page_size` (default 50, 1-100) as a config-driven per-request override
  (`instatusPageSize`/`instatusMaxPageSize`). The engine's `page_number` paginator reads its page
  size from `streams.json`'s static `base.pagination.page_size` field, not from a runtime
  `config.page_size` value (there is no dialect mechanism to wire a config value into
  `PaginationSpec.PageSize` at read time). This bundle declares `page_size: 50` — legacy's own
  default — as a fixed constant; a caller wanting a different page size (as legacy permitted via
  its `page_size` config key) cannot express that through this bundle. `spec.json` still declares
  `page_size` (default `"50"`) for documentation/informational parity, but no template anywhere in
  `streams.json` consumes it (F6-adjacent: kept because it documents the legacy config surface,
  unlike a genuinely-dead key with no legacy analog at all).
- **`max_pages` is not modeled.** Legacy's `instatusMaxPages` config-driven hard page-count cap has
  no equivalent `spec.json`/`streams.json` wiring in this bundle; pagination is bounded only by the
  short-page stop signal, matching every other page-count-unbounded stream in this dialect.
- Instatus's four core list resources are the only streams migrated; any additional
  write/mutation endpoints are out of scope for wave2 — see `api_surface.json`'s `excluded:
  {category: out_of_scope, reason: "Pass B capability expansion"}` entries.
