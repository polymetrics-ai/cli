# Overview

Harness is a wave2 fan-out declarative-HTTP migration. It reads Harness NextGen organizations,
projects, services, connectors, and pipelines through the Harness platform REST API
(`GET https://app.harness.io/...`). This bundle targets capability parity with
`internal/connectors/harness` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide the Harness account identifier via the `account_id` config value and a Harness NextGen API
key via the `api_key` secret; both are required. The API key is sent as the `x-api-key` header
(`mode: api_key_header`), matching legacy's `connsdk.APIKeyHeader("x-api-key", secret, "")` exactly
(no prefix). `base_url` defaults to `https://app.harness.io` and may be overridden for tests,
self-managed instances, or proxies. Legacy also accepts `account_identifier` and `api_url` as
fallback alias config keys; this bundle declares only the primary keys (`account_id`, `base_url`)
since a `spec.json` property with no template consuming it is dead config (conventions.md §3's
query-templating note, applied the same way to spec properties in general) — the aliases are
scope-narrowed out, documented here rather than declared-but-unwireable.

## Streams notes

All 5 streams (`organizations`, `projects`, `services`, `connectors`, `pipelines`) are `GET` list
endpoints scoped by the required `accountIdentifier` query parameter, matching legacy's
`harnessStreamEndpoints` table. `organizations`/`projects`/`services`/`connectors` wrap each
`data.content[]` element under a singular per-stream key (`organization`/`project`/`service`/
`connector`); `pipelines` does not (legacy's own `wrapper: ""`). This bundle reproduces the unwrap
with `computed_fields` reaching into `record.<wrapper>.<field>` (e.g.
`"identifier": "{{ record.organization.identifier }}"`), matching legacy's `unwrap()` helper
exactly — `records.path: "data.content"` extracts the wrapped elements, and `computed_fields`
resolves against the RAW (pre-projection) element, so the nested `organization`/`project`/etc.
object is still reachable. `pipelines`' `computed_fields` reference bare `record.<field>` (no
wrapper), matching `wrapper: ""`.

## Write actions & risks

None. Harness is a read-only source connector (`capabilities.write: false`): organization/project/
service listings have no safe reverse-ETL write semantics, matching legacy's own doc comment. This
bundle ships no `writes.json`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`.

## Known limits

- **`ENGINE_GAP`: 0-based page-index pagination cannot be expressed — pagination is scope-narrowed
  to a single page per stream.** Harness NextGen's list envelope is 0-indexed
  (`data.pageIndex`/`data.totalPages`, legacy's `harvest` loop starts at `page := 0` and requests
  `pageIndex=0,1,2,...` until `pageIndex+1 >= totalPages`). The engine's `page_number` pagination
  type (`PaginationSpec.StartPage`) cannot express a 0-based start: `engine/paginate.go`'s
  `newPaginator` unconditionally coerces `StartPage: 0` to `1`
  (`start := spec.StartPage; if start == 0 { start = 1 }`) — there is no way in this dialect to
  distinguish "explicitly declare page 0" from "start_page omitted, default to 1", since both are
  the Go zero value. Sending `pageIndex=1` as the FIRST request would silently skip the real first
  page of every Harness list response (verified against legacy's own test fixture:
  `harness_test.go`'s `TestReadPaginatesAndAuthenticates` serves org_1 at `pageIndex=0` and org_2
  at `pageIndex=1` — two genuinely different pages of data, not an off-by-one that happens to be
  harmless) — an accepted-input DATA change conventions.md §5's meta-rule forbids. The engine's
  `offset_limit` pagination type is equally unusable: it advances by a record OFFSET
  (`offset += PageSize` each round), not a page NUMBER, and Harness's `pageIndex` is a page index,
  not an offset — wiring `offset_param: pageIndex` would send `pageIndex=0,50,100,...` instead of
  `pageIndex=0,1,2,...`, which is wrong in the opposite direction. This bundle therefore declares
  `pagination.type: "none"` (a single request per stream, relying on the Harness API's own
  documented default of `pageIndex=0` when the param is omitted entirely — reproducing the FIRST
  page correctly) rather than risk silently skipping real data with a wrong start index. Every
  stream still sends `pageSize={{ config.page_size }}` (default `50`, matching legacy's
  `harnessDefaultPageSize`) to bound response size. **Result set beyond the first page is not
  read** — a genuine, documented scope-narrowing versus legacy's full pagination, not a silent
  divergence. This is a real `ENGINE_GAP`: a future increment letting `PaginationSpec` distinguish
  "start_page explicitly 0" from "start_page omitted" (e.g. a `*int` field, or a separate
  `zero_based: true` flag) would close it cleanly.
- **Legacy's fixture-mode-only behavior is not modeled.** Legacy's `readFixture` path (only
  reached when `config.mode == "fixture"`) emits a fixed synthetic shape (e.g. `modules:
  ["CD","CI"]` on every stream, not just `projects`) that diverges from the live per-stream record
  shapes. This bundle's schemas and fixtures target the live record shape only; the engine's own
  conformance/fixture-replay harness provides the credential-free test affordance legacy's fixture
  mode was built for.
- Full Harness API surface (pipeline execution, triggers, pipeline-execution-summary) is out of
  scope for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
