# Overview

k6 Cloud reads organizations, projects, and load tests through the k6 Cloud REST API
(`https://api.k6.io`). This bundle migrates all 3 of legacy `internal/connectors/k6-cloud`'s
streams to the declarative engine at capability parity, including `projects`, now expressed via the
engine's `stream.fan_out` dialect (see Streams notes). The legacy package stays registered and
unchanged until wave6's registry flip regardless. The connector is read-only (k6 Cloud load-test
resources have no obvious safe reverse-ETL writes, matching legacy).

## Auth setup

Provide a k6 Cloud API token via the `api_token` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_token>`) and is never logged.

## Streams notes

All 3 streams are ported at Tier 1, matching legacy's `k6StreamSpecs` entries for the same names
exactly:

- `organizations` (`GET /v3/organizations`) — no pagination (legacy's `paginated` flag is unset for
  this stream; the API returns every accessible organization in one response), full refresh,
  primary key `id`.
- `k6_tests` (`GET loadtests/v2/tests`) — `page_number` pagination (`page` starts at 1, request
  size driven by `query.page_size` from `config.page_size`, default `"32"` matching legacy's
  default). The pagination block's own short-page stop-threshold (`pagination.page_size: 32`) is a
  fixed literal matching the config default — see the pagination stop-threshold note below.
- `projects` (`GET /v3/organizations/{id}/projects`) — a sub-resource fan-out read, matching
  legacy's `readPerOrganization`/`collectOrganizationIDs`: the engine's `stream.fan_out` dialect
  first fully paginates `ids_from.request` (`GET /v3/organizations`, `records_path:
  "organizations"`, `id_field: "id"`) to resolve every accessible organization id ONCE per `Read()`
  call, then re-runs the stream's own request/pagination sequence once per resolved id with
  `into.path_var: "organization_id"` substituted for `{{ fanout.id }}` in `path`
  (`/v3/organizations/{{ fanout.id }}/projects`) — the identical `path_var` shape cisco-meraki's
  `organization_networks`/`organization_devices`/`organization_admins` and breezy-hr's `candidates`
  already use. **No `stamp_field` is declared**: unlike cisco-meraki (whose stamped
  `organizationId` field does not otherwise appear in the child payload, and is schema-typed
  `string` because every fan-out id resolves as a string), k6 Cloud's own `projects` API response
  already returns `organization_id` as a real integer field on every record, which legacy's
  `projectRecord` copies verbatim — declaring `stamp_field: "organization_id"` here would
  overwrite that wire-native integer with the fan-out id's string form post-projection, a real
  parity deviation (schema type `integer` vs. an engine-stamped `string`) rather than a neutral
  convenience. Schema projection alone reproduces the API's own `organization_id` value and type
  exactly, matching legacy exactly. Pagination within each organization's sub-sequence is
  `page_number`, identical shape to `k6_tests` (`page` starts at 1, `query.page_size` from
  `config.page_size`, default `"32"`) — legacy's `harvest` uses the same `pageSize`/`maxPages`
  config for every paginated stream, `projects` included.

All 3 streams' `mapRecord` functions in legacy are pure field-for-field copies with no renaming or
nesting to flatten, so no `computed_fields` are needed; schema projection alone reproduces legacy's
emitted shape exactly, including `k6_tests`' `test_run_ids` array field (copied by bare key match,
preserving its native JSON array type — draft-07 schema types it `["array", "null"]`).

## Write actions & risks

None. k6 Cloud is read-only in both legacy and this bundle (`capabilities.write: false`, no
`writes.json`) — legacy's own comment notes there are no obvious safe reverse-ETL targets for
load-test resources.

## Known limits

- **Stream name forced rename `k6-tests` -> `k6_tests` (ACCEPTABLE, forced deviation)**: legacy's
  stream name is literally `"k6-tests"` (with a hyphen). The engine's `streams.schema.json` enforces
  `name` matching `^[a-z][a-z0-9_]*$` for every stream (no hyphens allowed, unlike the connector's
  own directory/registry name, which does permit hyphens) — `connectorgen validate` hard-fails
  otherwise. This bundle's stream identifier is `k6_tests`; the raw API's own JSON response key
  (`"k6-tests"`, `streams.json`'s `records.path`) is unaffected and still matches the wire shape
  exactly — only the catalog-facing stream NAME changes, never any emitted record field or value.
- Full k6 Cloud API surface (test run details, insights, thresholds, etc.) is out of scope; see
  `api_surface.json` — only `organizations`/`k6_tests`/`projects` are implemented at Tier 1 in this
  wave; all 3 of legacy's streams are now covered.
- **Pagination stop-threshold parity narrowing (ACCEPTABLE, documented, same class as
  judge-me-reviews/just-sift)**: legacy's `page_size` config (1-100, default 32) drives both the
  `k6_tests`/`projects` request size and its own short-page stop check. The engine's `page_number`
  paginator's stop-threshold (`pagination.page_size`) is a fixed literal (`32`, legacy's default)
  that cannot be wired to the same runtime `config.page_size` value the request query param uses.
  Never wrong for the default-`page_size` case; only imprecise for a non-default override. This
  applies independently to each organization's `projects` sub-sequence, matching legacy's
  `harvest` using the same `pageSize` for every call regardless of which organization it's
  fetching.
- `organizations` fixture ships a single page (no pagination declared, matching legacy).
  `k6_tests` fixture ships a 32-record page 1 (matching the fixed `pagination.page_size: 32`
  short-page threshold) plus a 1-record page 2. `projects` fixture ships 3 pages: page 1 is the
  `fan_out.ids_from.request` organization-listing response (2 organizations, reusing the same
  fixture body as `organizations`' own page 1), and pages 2-3 are each organization's
  single-record (short-page, pagination-terminating) `projects` sub-request — the same 3-page
  shape cisco-meraki's `organization_networks`/`organization_devices`/`organization_admins` and
  breezy-hr's `candidates` fixtures already use for a `fan_out` stream.
