# Overview

CIMIS is a wave2 fan-out declarative-HTTP migration. It reads California Irrigation Management
Information System (CIMIS) weather station metadata through the CIMIS Web API
(`GET https://et.water.ca.gov/api/station`). This bundle targets capability parity with
`internal/connectors/cimis` (the hand-written connector it migrates) for the `stations` stream
only; the legacy package stays registered and unchanged until wave6's registry flip, and remains
authoritative for the `daily`/`hourly` streams this bundle does not cover (see Known limits).

## Auth setup

The `stations` stream (`GET /api/station`) is served by CIMIS without an appKey, matching legacy's
own comment ("the station endpoint is the lightest read that confirms connectivity; it does not
require the appKey"). `base.auth` is declared `[{"mode": "none"}]`. `api_key` is declared in
`spec.json` (`x-secret: true`, optional) as a reserved field for a future `daily`/`hourly` stream
expansion but is not wired into any template in this bundle — it currently has no effect.

## Streams notes

`stations` reads `GET /api/station`, records at the `Stations` top-level array key, primary key
`["StationNbr"]`, every field passed straight through unchanged (matching legacy's
`cimisStationRecord`, a verbatim passthrough with no renaming). No incremental cursor is published
(matching legacy: the `stations` stream carries no `CursorFields`).

## Write actions & risks

None. CIMIS is a read-only public data API (`capabilities.write: false`, no `writes.json`),
matching legacy's `Write` unconditionally returning `connectors.ErrUnsupportedOperation`.

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
  `spec.json` but not wired into any `auth` candidate in this bundle, since the only implemented
  stream (`stations`) never sends it. It remains reserved for when `daily`/`hourly` are added in a
  follow-up wave.
- `base_url`'s SSRF-guard scheme/host validation (https/http only, host required) is reproduced by
  the engine's own base-URL handling; no bundle-level behavior change.
