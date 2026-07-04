# Overview

Senseforce is a wave2 fan-out declarative-HTTP migration. It reads rows from a single configured
Senseforce dataset through the Senseforce Public API (`GET
{{ config.backend_url }}/api/v1/datasets/{{ config.dataset_id }}/records`). This bundle is
migrated from `internal/connectors/senseforce` (the hand-written connector it replaces); the
legacy package stays registered and unchanged until wave6's registry flip. Read-only
(`capabilities.write` is `false`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`).

## Auth setup

Provide a Senseforce API access token via the `access_token` secret; it is sent as a Bearer token
(`mode: bearer`), matching legacy's `connsdk.Bearer(token)`. `backend_url` (your Senseforce tenant
base URL, e.g. `https://yourtenant.senseforce.io`) and `dataset_id` (the dataset to read) are both
required config with no default, matching legacy's own required `backend_url`/`dataset_id`
validation (legacy has no tenant-derivable default either).

## Streams notes

`records` (`GET /api/v1/datasets/{{ config.dataset_id }}/records`, records at `data`) is the sole
stream, matching legacy's single hard-coded `"records"` stream name. Pagination is `page_number`
(`page`/`page_size`, `start_page: 1`, `page_size: 100`), matching legacy's
`connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "page_size", StartPage: 1, PageSize:
pageSize}` at its default `page_size` value (100 — legacy's `defaultPageSize`). Primary key is
`["id"]`. Legacy's `Catalog()` also names a `Timestamp` field as a `CursorFields` entry, but
`Read()` itself never applies any incremental/cursor filtering (it only paginates); this bundle
matches the actually-exercised behavior (full pagination, no `incremental` block) rather than the
unused `CursorFields` catalog hint — see Known limits.

## Write actions & risks

None. Legacy `senseforce.go`'s `Write` returns `connectors.ErrUnsupportedOperation`
unconditionally; `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`x-cursor-field` is intentionally NOT declared on `records`, even though legacy's `Catalog()`
  lists `Timestamp` under `CursorFields`.** Legacy's `Read` path never reads or filters on any
  cursor value — it always performs a full paginated sweep of the dataset regardless of any prior
  state. Declaring `x-cursor-field: Timestamp` here without a matching `incremental` block in
  `streams.json` would be inert (the engine only performs incremental filtering when a stream
  declares `incremental`), and adding a real `incremental` block would be new behavior legacy never
  had (Senseforce's dataset-records endpoint accepts no time-range query parameter to filter
  server-side, and `Timestamp` field values are not guaranteed monotonic across the raw dataset).
  Full-refresh-only parity is therefore the correct, honest representation.
- **`page_size`/`max_pages` are NOT exposed as runtime-configurable `spec.json` properties**, even
  though legacy accepts both via `req.Config.Config["page_size"]`/`["max_pages"]`
  (`positiveInt`/`parseMaxPages`, 1-1000 / non-negative integer respectively). `PaginationSpec`'s
  `PageSize`/`MaxPages` fields (`internal/connectors/engine/bundle.go`) are plain JSON
  ints — the dialect has no templating mechanism for either field, so a `spec.json` property
  feeding them would be structurally impossible to wire. A `page_number` paginator's own
  per-page query (which carries `size_param`'s resolved value, sourced only from the static
  `PaginationSpec.PageSize`) is additionally regenerated fresh on every page via `read.go`'s
  `mergeQuery(baseQuery, page.Query)`, where `page.Query` always overwrites any same-keyed
  `stream.Query` entry — so even a `stream.Query["page_size"]` entry pointed at
  `{{ config.page_size }}` would be silently discarded every request, never actually reaching the
  wire. This bundle therefore hardcodes `page_size: 100`/`max_pages: 1` (legacy's own default
  values) in `streams.json`'s `base.pagination` block and declares neither as spec config —
  declaring a config property no template anywhere in the bundle can consume would be dead config
  (F6, conventions.md), not a genuine override. This never changes which records are emitted for
  any caller that (like most operators) never overrides Senseforce's own defaults; a caller that
  legitimately needs a non-default `page_size`/`max_pages` is out of scope for this wave.
- **Pass B full-surface review result: `already_full`.** Senseforce's ENTIRE documented public API
  (`manual.senseforce.io/manual/sf-platform/public-api/endpoints`) is a single element-by-id
  request family: "get the results of predefined datasets" by a dataset id taken from the
  dashboard URL, with an optional `DatasetFilters` body for column-scoped filtering. There is no
  documented dataset-listing/discovery endpoint, no dataset-metadata/column-schema endpoint, and no
  write/mutation endpoint of any kind anywhere in the manual — the public API exists to let
  external tools pull already-built dataset results out of the platform, not to manage platform
  configuration, so `capabilities.write: false` is not a scope narrowing but an accurate reflection
  of the real API's shape. The ONLY other documented element type is a separate script-execution
  endpoint (`POST /api/v1/scripts/{script_id}/results`, accepting its own `ScriptFilters` body) —
  a fundamentally different resource (a published Senseforce SCRIPT, not a dataset), excluded via
  `api_surface.json` as `out_of_scope` rather than modeled as a second stream, since it would need
  its own `script_id` config surface and has no relationship to the `dataset_id`-scoped `records`
  stream this bundle already fully covers. See `api_surface.json`'s `excluded` entry for the exact
  reasoning.
