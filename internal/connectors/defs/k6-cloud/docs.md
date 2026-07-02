# Overview

k6 Cloud reads organizations and load tests through the k6 Cloud REST API
(`https://api.k6.io`). This bundle migrates 2 of legacy `internal/connectors/k6-cloud`'s 3 streams
to the declarative engine at capability parity; `projects` remains on the legacy Go implementation
(see Known limits — a genuine Tier-2 sub-resource fan-out, out of scope for this wave). The legacy
package stays registered and unchanged until wave6's registry flip regardless. The connector is
read-only (k6 Cloud load-test resources have no obvious safe reverse-ETL writes, matching legacy).

## Auth setup

Provide a k6 Cloud API token via the `api_token` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_token>`) and is never logged.

## Streams notes

Two streams are ported at Tier 1, matching legacy's `k6StreamSpecs` entries for the same names
exactly:

- `organizations` (`GET /v3/organizations`) — no pagination (legacy's `paginated` flag is unset for
  this stream; the API returns every accessible organization in one response), full refresh,
  primary key `id`.
- `k6_tests` (`GET loadtests/v2/tests`) — `page_number` pagination (`page` starts at 1, request
  size driven by `query.page_size` from `config.page_size`, default `"32"` matching legacy's
  default). The pagination block's own short-page stop-threshold (`pagination.page_size: 32`) is a
  fixed literal matching the config default — see the pagination stop-threshold note below.

Both streams' `mapRecord` functions in legacy are pure field-for-field copies with no renaming or
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
- **`projects` (`GET /v3/organizations/{id}/projects`) is NOT ported in this wave — kept on the
  legacy Go implementation.** Legacy's `readPerOrganization` reads this stream by first listing
  every accessible organization id (`collectOrganizationIDs`, a call to the `organizations`
  endpoint), then issuing one paginated `GET /v3/organizations/{id}/projects` request PER
  organization id and aggregating every organization's projects into one stream. This is a genuine
  sub-resource fan-out read (the github issue→comments pattern) — `docs/migration/conventions.md`
  section 1 explicitly lists "sub-resource fan-out reads" as a Tier-2 trigger requiring a
  `StreamHook` (`ReadStream(ctx, stream, req, rt, emit) (handled bool, err error)`); there is no
  Tier-1 JSON mechanism to (a) read one endpoint's response to discover a set of ids and (b) issue
  one paginated sub-request per discovered id. Per this wave's hard rule (Tier-2/Tier-3 escape
  hatches are out of scope for fan-out migration agents — a follow-up wave implements hooks), this
  bundle creates no Go hook and leaves `projects` on the legacy connector.
- Full k6 Cloud API surface (test run details, insights, thresholds, etc.) is out of scope; see
  `api_surface.json` — only `organizations`/`k6_tests` are implemented at Tier 1 in this wave.
- **Pagination stop-threshold parity narrowing (ACCEPTABLE, documented, same class as
  judge-me-reviews/just-sift)**: legacy's `page_size` config (1-100, default 32) drives both the
  `k6_tests` request size and its own short-page stop check. The engine's `page_number` paginator's
  stop-threshold (`pagination.page_size`) is a fixed literal (`32`, legacy's default) that cannot be
  wired to the same runtime `config.page_size` value the request query param uses. Never wrong for
  the default-`page_size` case; only imprecise for a non-default override.
- `organizations` fixture ships a single page (no pagination declared, matching legacy).
  `k6_tests` fixture ships a 32-record page 1 (matching the fixed `pagination.page_size: 32`
  short-page threshold) plus a 1-record page 2.
