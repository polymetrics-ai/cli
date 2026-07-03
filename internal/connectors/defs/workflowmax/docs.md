# Overview

WorkflowMax is a wave2 fan-out declarative-HTTP migration. It reads jobs, clients, and contacts
through the WorkflowMax API (`GET {{ config.base_url }}/...`). This bundle is migrated from
`internal/connectors/workflowmax` (the hand-written connector it replaces); the legacy package
stays registered and unchanged until wave6's registry flip. Read-only (`capabilities.write` is
`false`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`).

## Auth setup

Provide a WorkflowMax API access token via the `access_token` secret; it is sent as a Bearer
token on every request (`mode: bearer`), matching legacy's `connsdk.Bearer(token)`. `base_url`
defaults to `https://api.workflowmax2.com` (legacy's `defaultBaseURL`) and may be overridden for
test proxies.

## Streams notes

All 3 streams (`jobs`, `clients`, `contacts`) share the identical envelope (records at the
top-level `data` array) and `page_number` pagination (`page`/`page_size` query params, matching
legacy's `PageNumberPaginator{PageParam: "page", SizeParam: "page_size", StartPage: 1}`). The
base pagination block's `page_size: 100` mirrors legacy's `defaultPageSize`; legacy bounds it to
a max of 500 (`maxPageSize`) and `max_pages` defaults to 1 (legacy's `readMaxPages` default) when
unset — a config-driven `page_size`/`max_pages` override is not modeled (see Known limits).
`jobs` declares a stream-level `pagination` override (`page_size: 2`) so its conformance fixture
(`fixtures/streams/jobs/{page_1,page_2}.json`) can ship a genuine full page-1 (2 records) followed
by a genuine short page-2 (1 record) that proves real two-page termination
(`docs/migration/conventions.md` §4) — this is a fixture-authoring device, not a behavior
difference from `clients`/`contacts`, which share the base 100-per-page default and ship a single
fixture page (2 records, an honestly short final page under a 100 page-size, matching legacy's
own default).

`jobs` (`GET /jobs`) emits `id`/`name`/`updated_at`, matching legacy's field set exactly.
`clients` (`GET /clients`) emits the identical shape. `contacts` (`GET /contacts`) additionally
emits `email`. Primary key is `id` for every stream; `updated_at` is declared as the incremental
cursor field for manifest-surface parity, matching legacy's `cursorFields`, though neither legacy
nor this bundle actually issues a server-side incremental filter — legacy's `Read` performs a
full stream read every time regardless of any prior cursor.

## Write actions & risks

None. Legacy `workflowmax.go`'s `Write` returns `connectors.ErrUnsupportedOperation`
unconditionally; `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` config-driven overrides are not modeled.** Legacy reads
  `config["page_size"]` (bounded 1-500) and `config["max_pages"]` (default 1) at request time via
  `boundedInt`/`readMaxPages`. The engine's `page_number` paginator reads `PaginationSpec.PageSize`
  from the static `streams.json` `base.pagination` block only — there is no per-request
  config-driven override mechanism for either value in the current dialect (matching the same gap
  documented for other page-number-paginated wave2 bundles, e.g. `docs/migration/conventions.md`
  §3's optional-query-dialect discussion). `page_size`/`max_pages` remain declared in `spec.json`
  as documentation of legacy's accepted config surface, but neither is wired into any template in
  this bundle.
- **No incremental filter is modeled**, matching legacy: `updated_at` is declared as
  `x-cursor-field` for manifest parity, but WorkflowMax's `/jobs`, `/clients`, and `/contacts`
  endpoints (as legacy calls them) accept no time-range query parameter — both connectors always
  perform a full stream read on every sync.
- The full WorkflowMax API surface (job/client/contact mutation, invoicing, timesheets) is out of
  scope for this wave; see `api_surface.json`'s `excluded` entries.
