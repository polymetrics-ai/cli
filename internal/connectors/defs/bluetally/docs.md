# Overview

BlueTally is a wave2 fan-out declarative-HTTP migration. It reads BlueTally IT asset management
data (assets, employees, licenses, maintenances, accessories) through the BlueTally REST API
(`GET https://app.bluetallyapp.com/api/v1/<resource>`). This bundle is migrated from
`internal/connectors/bluetally` (the hand-written connector); the legacy package stays registered
and unchanged until wave6's registry flip.

## Auth setup

Provide a BlueTally API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`), matching legacy's `connsdk.Bearer(secret)`
(`bluetally.go:275`). Never logged. `base_url` defaults to `https://app.bluetallyapp.com` and may
be overridden for tests/proxies.

## Streams notes

All five streams (`assets`, `employees`, `licenses`, `maintenances`, `accessories`) read
`/api/v1/<resource>` list endpoints, each returning a bare JSON array at the response root
(`records.path: ""`), matching legacy's root-path `connsdk.RecordsAt(resp.Body, "")` convention.
Pagination is `offset_limit` (`limit`/`offset` query params, `page_size: 50` — legacy's own
`bluetallyDefaultPageSize`); the next page's `offset` advances by the page size and the engine
stops on a short/empty page, identical to legacy's own `len(records) < pageSize` stop rule and
`offset = page * pageSize` request construction.

Every stream declares `updated_at` as `x-cursor-field` for manifest-surface parity with legacy's
`CursorFields: []string{"updated_at"}` on every stream, but — like legacy — none of them actually
filter server-side: legacy's `harvest` never sends any lower-bound query parameter (BlueTally's
list API supports only offset/limit pagination, no time-based filter), so every read is a full
sync regardless of a cursor's value. This bundle declares no `incremental` block on any stream,
matching legacy's behavior exactly (a declared cursor field with no server-side filter or
client-side dedup is legacy's own shape, not a migration gap).

## Write actions & risks

None. BlueTally is read-only in legacy (the upstream source supports full-refresh reads only);
`capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's `Write`
returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size`
  (1-100, default 50) and `max_pages` (0/all/unlimited default) as config-driven overrides
  (`bluetallyPageSize`/`bluetallyMaxPages`, `bluetally.go:314-342`). The engine's `offset_limit`
  paginator's `page_size` is a fixed value baked into `streams.json`'s `base.pagination` block at
  bundle-author time (`PaginationSpec.PageSize` is a plain int, never `config.*`-templated), and
  `MaxPages` is likewise a fixed bundle-time int — matching the identical, already-documented
  searxng/bitly precedent (`docs/migration/conventions.md`). This bundle bakes in legacy's own
  default (`page_size: 50`); `max_pages` is left unbounded (0/omitted), matching legacy's own
  default (empty `max_pages` config = unlimited). Neither is declared in `spec.json` (F6,
  REVIEW.md: a declared-but-unwireable config key is worse than an absent one).
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only
  reached when `config.mode == "fixture"`) stamps a `previous_cursor` field (echoing
  `req.State["cursor"]` when set) onto every fixture-mode record (`bluetally.go:243-247`), and
  synthesizes a superset of fields shared loosely across all five streams rather than each
  stream's own real wire shape. This is not part of the LIVE record shape; this bundle's schemas
  and fixtures instead use each stream's real per-resource wire shape from the live `harvest`
  path, matching each stream's own `bluetally*Record`/`bluetally*Fields` mapping in
  `internal/connectors/bluetally/streams.go`.
