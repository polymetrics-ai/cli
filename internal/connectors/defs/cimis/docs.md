# Overview

CIMIS is a Pass B full-surface declarative-HTTP connector. It reads California Irrigation
Management Information System (CIMIS) weather station metadata and both documented zip-code
reference lists through the CIMIS Web API (`GET https://et.water.ca.gov/api/{station,
stationzipcode,spatialzipcode}`). This bundle targets capability parity with
`internal/connectors/cimis` (the hand-written connector it migrates) for the `stations` stream, and
goes beyond legacy by adding the two zip-code reference streams legacy never implemented (`GET
/api/stationzipcode`, `GET /api/spatialzipcode` — both genuinely Tier-1 expressible, simple
top-level-array-under-one-key GET resources, no appKey required); the legacy package stays
registered and unchanged until wave6's registry flip, and remains authoritative for the
`daily`/`hourly` streams this bundle does not cover (see Known limits).

## Auth setup

All 3 streams (`stations`, `station_zip_codes`, `spatial_zip_codes`) are served by CIMIS without an
appKey — CIMIS's own docs state registration/an appKey is required only for the weather/ET data
services (`/api/data`), never for station or zip-code metadata. `base.auth` is declared
`[{"mode": "none"}]`. `api_key` is declared in `spec.json` (`x-secret: true`, optional) as a reserved
field for a future `daily`/`hourly` stream expansion but is not wired into any template in this
bundle — it currently has no effect.

## Streams notes

`stations` reads `GET /api/station`, records at the `Stations` top-level array key, primary key
`["StationNbr"]`, every field passed straight through unchanged via `"projection": "passthrough"`
(matching legacy's `cimisStationRecord`, which copies every key from the raw item onto the record
with no allowlist and no renaming: `for key, value := range item { rec[key] = value }`). No
incremental cursor is published (matching legacy: the `stations` stream carries no `CursorFields`).

`station_zip_codes` reads `GET /api/stationzipcode` (records at the `ZipCodes` top-level array key),
returning which CIMIS Weather Station Network stations serve which US zip codes — fields
`StationNbr`/`ZipCode`/`ConnectDate`/`DisconnectDate`/`IsActive`, primary key
`["StationNbr", "ZipCode"]` (a station can serve, and a zip code can be served by, more than one of
the other — the compound key is the documented uniqueness constraint for this resource).
`spatial_zip_codes` reads `GET /api/spatialzipcode` (same `ZipCodes` envelope shape, one level
narrower: no `StationNbr` field at all, since Spatial CIMIS System zip-code coverage isn't tied to
a physical station), primary key `["ZipCode"]`. Both new streams use `"projection": "passthrough"`,
matching `stations`' own no-allowlist, no-renaming shape (there is no legacy record mapper to match
parity against, since legacy never implemented either resource) and CIMIS's own flat, already
schema-friendly field naming. Neither publishes an incremental cursor: CIMIS's zip-code reference
data is a small, slowly-changing lookup table with no documented updated-at/modified-since filter.

## Write actions & risks

None. CIMIS is a read-only public reference/observational data API — every one of its documented
endpoints is a `GET`; there is no create/update/delete surface anywhere in the CIMIS Web API to
model as a write action (`capabilities.write: false`, no `writes.json`), matching legacy's `Write`
unconditionally returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`daily` and `hourly` are not modeled as streams in this bundle (ENGINE_GAP).** Both legacy
  streams read `GET /api/data`, whose records live nested at `Data.Providers[].Records[]` — an
  array WITHIN an array, requiring flattening across a dynamic number of provider entries
  (`flattenProviderRecords`, `cimis.go:232-257`). The engine's declarative record extraction
  (`connsdk.RecordsAt`, `internal/connectors/connsdk/extract.go:33`) walks a dotted path through
  object KEYS only (`selectPath`); it cannot descend into an array (`Providers`) to reach a further
  nested array (`Records`) inside each element, so there is no `records.path` value that can even
  locate the array of interest, let alone flatten N providers' worth into one record set. Separately,
  each record's data-item fields arrive as a DYNAMIC, config-driven set of nested `{Value,Qc,Unit}`
  objects (which items appear depends on the `dataItems` config value sent on the request; e.g.
  `DayAirTmpAvg`, `DayEto`, `DayPrecip`, ...), and legacy's `cimisDataRecord`
  (`streams.go:104-120`) flattens EVERY such key present into 3 synthetic sibling fields
  (`<Item>_Value`/`<Item>_Qc`/`<Item>_Unit`) while also preserving the original nested object. The
  engine's `computed_fields` dialect requires statically-named output fields declared once in
  `streams.json` — there is no "for every key matching this dynamic shape, synthesize N sibling
  fields" primitive. Both gaps are genuine `ENGINE_GAP`s, not Tier-2-fixable shapes (no single hook
  interface flattens the record stream itself without also reimplementing the whole read loop, which
  is a StreamHook — forbidden this wave per the fan-out task's hard rules). Legacy stays
  authoritative for `daily`/`hourly` until the engine gains both a nested-array-within-array
  flatten primitive and a dynamic-key flatten primitive.
- `api_key`/appKey query-param auth (`connsdk.APIKeyQuery("appKey", secret)`) is declared in
  `spec.json` but not wired into any `auth` candidate in this bundle, since none of the 3 implemented
  streams ever sends it. It remains reserved for when `daily`/`hourly` are added in a follow-up wave.
- **`/api/station/{stationNumber}`, `/api/stationzipcode/{zipCode}`, and
  `/api/spatialzipcode/{zipCode}`** (single-item path-parameter filters of the 3 covered list
  resources) are excluded as `duplicate_of` rather than modeled as additional streams: a full list
  read of the parent resource already returns every record the filtered variant would return
  individually, and the engine's declarative stream dialect has no per-record "read one item by path
  parameter" primitive distinct from an ordinary list stream.
- **`station_zip_codes`' `StationNbr` schema type is `["integer", "string"]`, not a tightened bare
  `integer`.** CIMIS's own REST reference example shows this field as a bare JSON integer (unlike the
  `stations` stream's `StationNbr`, which is a JSON string in that resource), but this bundle's
  `station_zip_codes` fixture was authored from documentation rather than a captured live response for
  this specific resource; the union type is an honest hedge against CIMIS's broader documented
  tendency to emit numeric-looking fields as strings elsewhere in this same API (`stations`' own
  `StationNbr`, `Elevation`, etc. are all strings). `"projection": "passthrough"` means the schema
  type here is documentation of the expected wire shape, not an enforced coercion — whichever type
  CIMIS actually sends survives unchanged either way.
- `base_url`'s SSRF-guard scheme/host validation (https/http only, host required) is reproduced by
  the engine's own base-URL handling; no bundle-level behavior change.
