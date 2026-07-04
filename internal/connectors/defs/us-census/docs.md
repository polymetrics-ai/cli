# Overview

US Census is a Tier-2 (StreamHook) migration, originally quarantined in wave1 under `ENGINE_GAP`
(`docs/migration/quarantine.json`): legacy's sole `query` stream returns a raw top-level JSON
**array-of-arrays** — a header row of caller-driven field names, followed by data rows — with
field names derived entirely at read-time from that header row, itself a function of the
caller-supplied `query_path`/`query_params` (`get=...`) qualifier. There is no fixed schema the
declarative `records`/`schema` dialect can project, matching the recorded blocker reason verbatim.
This bundle resolves the blocker via `hooks/us-census/hooks.go`'s `StreamHook`, porting legacy
`us-census/us_census.go`'s `censusRows` header-mapping logic almost verbatim. Read-only: legacy's
`Write` always returns `ErrUnsupportedOperation`, and this bundle declares `capabilities.write:
false` with no `writes.json` to match.

This revision is a Pass B full-surface review: the live API was re-fetched directly
(`https://api.census.gov/data.json`) to confirm the real, current surface. The caller-driven
`query` endpoint remains the only per-dataset data-query resource (unchanged from wave1) — but the
Bureau also publishes its own dataset discovery catalog at a FIXED, uniform-shaped endpoint
(`GET /data.json`), which this revision adds as a new `datasets` stream — the one genuinely new
declaratively-shapeable resource this API offers beyond the caller-configured query itself.

## Auth setup

Auth is genuinely simple and needs no hook at all: legacy sends the configured `api_key` secret as
a `key` query-string parameter on every request (`connsdk.APIKeyQuery("key", ...)`,
`us_census.go:116`). `streams.json`'s `base.auth` declares this declaratively —
`{"mode": "api_key_query", "param": "key", "value": "{{ secrets.api_key }}", "when": "{{
secrets.api_key }}"}` falling back to `{"mode": "none"}` when `api_key` is unset — matching
legacy's own tolerance of an empty key (legacy never validates `api_key` is non-empty; an unset key
sends `key=` with no value, which the live Census API would itself reject, not the connector).

## Streams notes

One stream, `query`. `hooks/us-census/hooks.go`'s `StreamHook.ReadStream` always returns
`handled=true`: it builds the request path/query from `config.query_path`/`config.query_params`
exactly like legacy's `Read` (`url.ParseQuery` on `query_params`, a GET to `query_path` with the
resolved `key` query param via `rt.Requester`), decodes the response as `[][]any` (mirroring
`censusRows`'s `json.NewDecoder(...).UseNumber()` decode), lower-cases the header row's values into
field names, and emits one record per data row with those dynamic keys. A response with fewer than
2 rows (no header, or header-only) emits zero records — matching legacy's `if len(rows) < 2 {
return nil, nil }` early return exactly. `mode: fixture` short-circuits to a single canned
`{"name": "United States", "estab": "1"}` record, matching legacy's `emitFixture` verbatim (test/
conformance-harness affordance only, never set in production).

`query`'s declarative `path`/`records`/`schema` fields are an inert shadow, never exercised
in production dispatch (which always routes through the StreamHook) — see this stream's own
`conformance.skip_dynamic` marker in `streams.json` (stream-level, not bundle-level: `datasets`,
below, is an ordinary declarative stream with no hook involvement and gets full dynamic conformance
coverage). `schemas/query.json` declares `additionalProperties: true` and the two illustrative
fields (`name`/`estab`) legacy's own catalog `Fields` list documents, since the real field set is
unbounded and caller-driven.

**No incremental sync mode on `query`**: legacy's `query` stream declares no cursor field and no
pagination — every call re-reads the full configured query. This bundle matches: no `incremental`
block, no `pagination` beyond the base's `{"type": "none"}`.

**Pass B addition — `datasets`** (`GET data.json`, records at the top-level `dataset` array): the
Census Bureau's own dataset-discovery catalog, confirmed live during this review (a DCAT-formatted
JSON document, NOT an array-of-arrays like `query` — this is a genuinely different, fixed-shape
resource, so it needs no hook and is modeled as a plain declarative stream). Each row's real wire
shape nests the dataset's path segments as a JSON array under `c_dataset` (e.g. `["cbp"]` for a
`data/2019/cbp`-style path, sibling to a top-level `c_vintage` year field) — `computed_fields`
joins that array into a single `dataset_path` convenience field
(`{{ record.c_dataset | join:/ }}`) a caller could feed directly into this bundle's own
`query_path` config value for a different connection. No pagination (the live catalog is one JSON
document, not a paginated list) and no `incremental` block (the catalog is a slowly-changing
reference list, not a stream of new records — `modified` exists per-row but is dataset-level
metadata, not a sync cursor this stream tracks).

## Write actions & risks

None — US Census is read-only. `capabilities.write: false`, no `writes.json` file, matching
legacy's `ErrUnsupportedOperation` (`us_census.go:107-109`). The Census Bureau's public API is
itself read-only (a government open-data publishing API, not a data platform with a write surface).

## Known limits

- **Dynamic, caller-driven schema is inherent to the API, not a migration shortcut.** Every record's
  actual field set depends on the `get=` qualifier the operator configures in `query_params`; this
  bundle's `schemas/query.json` cannot enumerate a closed field list and instead declares
  `additionalProperties: true` alongside legacy's own two illustrative fields. This is the
  documented resolution of the recorded `ENGINE_GAP` blocker (`docs/migration/quarantine.json`), not
  a new deviation.
- **`query_path`/`query_params` are per-connection config, not per-stream** — matching legacy
  exactly (`us_census.go:80-87`): both are required config keys read once per `Read()` call, not
  templated per-stream. A single connection can therefore only ever sync one configured Census
  query at a time (legacy's own scope, unchanged).
- Rows whose header cell is empty, or whose row is shorter than the header, silently skip that
  column for that record (`censusRows`'s `if i >= len(headers) || headers[i] == "" { continue }`) —
  ported verbatim in the hook, not a new deviation.
- **Per-dataset sub-resources (`variables.json`/`geography.json`/`tags.json`/`groups.json`/
  `examples.json`) are out of scope.** The live catalog embeds one distinct URL per published
  dataset (thousands of datasets, e.g. `c_variablesLink`) for each of these sub-resource types —
  there is no fixed, finite endpoint to enumerate the way `datasets` itself is one fixed URL;
  covering them would require a `fan_out`-style stream keyed on every catalog row's own link field,
  which was judged out of scope for this pass (see `api_surface.json`'s `excluded` entries). A
  future capability-expansion pass could add a `fan_out` stream for `variables.json` specifically
  if a concrete need for column-metadata discovery arises.
- `datasets` shares the SAME base `auth`/`base_url` as `query` (this bundle has no per-stream auth
  override mechanism) — sending the configured `api_key` as a `key` query param to `/data.json` is
  harmless (the live catalog endpoint does not require authentication and silently ignores an
  unrecognized query param), not a behavior concern.
