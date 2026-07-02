# Overview

Statsig is a wave2 fan-out declarative-HTTP migration. It reads Statsig feature gates, dynamic
configs, experiments, and segments through the Statsig Console API
(`GET https://statsigapi.net/console/v1/...`). This bundle targets capability parity with
`internal/connectors/statsig` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Statsig Console API key via the `api_key` secret; it is sent as the `STATSIG-API-KEY`
header with no prefix, matching legacy's `connsdk.APIKeyHeader("STATSIG-API-KEY", key, "")`
(`statsig.go:114`). Never logged. `base_url` defaults to `https://statsigapi.net/console/v1` and
may be overridden for tests/proxies.

## Streams notes

All 4 streams (`feature_gates`, `dynamic_configs`, `experiments`, `segments`) are `GET` list
endpoints (`/gates`, `/dynamic_configs`, `/experiments`, `/segments`) sharing the identical record
shape: records live at the `data` key, primary key `["id"]`, and each record's fields (`id`, `name`,
`description`, `status`, `isEnabled`) are copied verbatim, matching legacy's shared
`copyRecord("id", "name", "description", "status", "isEnabled")` mapper exactly. None of the 4
streams declare a cursor field, matching legacy (no `CursorFields` on any of the 4 `connectors.Stream`
declarations — full refresh only).

Pagination is offset+limit (`pagination.type: offset_limit`, `limit_param: limit`, `offset_param:
offset`, `page_size: 100`), identical to legacy's `connsdk.OffsetPaginator{LimitParam: "limit",
OffsetParam: "offset", PageSize: pageSize}` with legacy's default `pageSize` of 100.

## Write actions & risks

None. Legacy's `Write` always returns `connectors.ErrUnsupportedOperation`; `capabilities.write` is
`false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`statsig.go`'s `pageSize`/`maxPages`, bounded 1-1000 / a non-negative integer or
  `all`/`unlimited`). The engine's `offset_limit` paginator has no config-driven page-size or
  max-pages knob (`PaginationSpec.PageSize`/`MaxPages` are static bundle JSON, never templated), so
  this bundle uses legacy's own default (`page_size: 100`) as a fixed bundle value and does not
  declare `page_size`/`max_pages` in `spec.json` at all (a declared-but-unwireable config key is
  worse than an absent one, per `docs/migration/conventions.md` F6). Pagination is unbounded by
  default (reads every page until a short page), matching legacy's own default of `maxPages=0`
  (unbounded) when `max_pages` is unset.
- Full Statsig Console API surface (mutation endpoints, metrics, target apps, audit logs) is out of
  scope for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
