# Overview

Adjust is a Tier-1 declarative-HTTP migration (catalog-labeled native/destination — that label is
WRONG: legacy `internal/connectors/adjust/adjust.go` is a `connsdk.Requester`-based, read-only
report reader with no write actions and no protocol-native/SDK dependency, so it belongs at
Tier 1, not Tier 3 or a destination). It reads Adjust Report Service report rows for configured
dimensions and metrics through the documented `reports-service/report` endpoint. This bundle is
engine-vs-legacy parity-tested against `internal/connectors/adjust` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide an Adjust Report Service API token via the `api_token` secret; it is used only for Bearer
auth (`Authorization: Bearer <api_token>`) and is never logged.

## Streams notes

The single `reports` stream issues `GET /reports-service/report` with `dimensions` and `metrics`
query parameters (each a comma-separated list, defaulting to `country` and `installs`
respectively — matching legacy's `csvOrDefault` defaults). An optional `additional_metrics` config
value is sent verbatim when set, omitted entirely otherwise (`omit_when_absent`). When both
`start_date` and `end_date` are configured, a `date_period` query parameter is sent as
`{{ start_date }}:{{ end_date }}` (legacy's `date_period` range format); when either is unset, the
parameter is omitted entirely rather than sent malformed (also `omit_when_absent` — the underlying
template hard-errors on either absent key, and the object form's absence-tolerance catches it).

Pagination follows Adjust's `next_page` convention: `pagination.type: cursor` with
`token_path: next_page` reads the next page number straight from the response body and resends it
as the `page` query parameter; pagination stops when `next_page` is absent or empty, matching
legacy's `strings.TrimSpace(next) == ""` stop condition exactly.

Each raw report row carries the row's dimension/metric values nested under `dimensions`/`metrics`
sub-objects (Adjust's real wire shape); legacy's `reportRecord` flattens both objects' keys onto
the top-level record and drops the `dimensions`/`metrics` wrapper keys themselves. This bundle
reproduces that flattening via `computed_fields`, each of which reaches into the raw nested path
(`record.dimensions.date`, `record.metrics.installs`, etc.) and schema-mode projection then keeps
only those declared top-level fields — the identical effective output legacy produces for any
configuration using the 6 fields below.

## Write actions & risks

None. This is a read-only source connector — `capabilities.write` is `false` and `Write` always
returns `ErrUnsupportedOperation`, matching legacy exactly.

## Known limits

- Legacy's `reportRecord` flattens **every** key found under the raw `dimensions`/`metrics`
  objects, whatever dimensions/metrics the caller configured (fully dynamic w.r.t.
  `cfg.Config["dimensions"]`/`cfg.Config["metrics"]`). This bundle's `computed_fields` instead
  declares a **fixed** set of 6 flattened fields (`date`, `country`, `app`, `installs`, `clicks`,
  `cost`) — the same fixed field set legacy's own `Catalog()` advertises. A caller who configures a
  dimension/metric outside this set (e.g. `os_name` or `revenue`) will have that field silently
  dropped by schema projection here, whereas legacy would have emitted it dynamically. This is a
  documented scope narrowing (the dialect's `computed_fields` map is declared statically per
  stream, with no mechanism to flatten an arbitrary caller-configured key set) — not a data
  distortion for the catalog-declared default configuration (`dimensions=country`,
  `metrics=installs`), which is fully reproduced.
- Legacy's `reportQuery` also recognized `ingest_start`/`until` as aliases for `start_date`/
  `end_date`. This bundle's `spec.json` only declares `start_date`/`end_date` — a `spec.json`
  property with no template consuming it is dead config (per this repo's dead-config rule), and a
  second alias pair adds no expressive power the primary pair lacks. Both mean lower/upper Adjust
  report period bounds.
- Adjust's campaign/creative-set management and KPI-service endpoints are not implemented (legacy
  never implemented them either); see `api_surface.json`'s scope note.
