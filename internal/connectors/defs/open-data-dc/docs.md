# Overview

Open Data DC is a read-only declarative-HTTP migration (wave2 fan-out) of
`internal/connectors/open-data-dc` (the hand-written connector it replaces at capability parity). It
reads District of Columbia Master Address Repository (MAR 2) locations, units, and SSL parcel records
through the public Open Data DC API. The legacy package stays registered and unchanged until wave6's
registry flip.

## Auth setup

Provide the MAR 2 API key as the `api_key` secret; it is sent as the `apikey` query parameter
(`{"mode": "api_key_query", "param": "apikey", "value": "{{ secrets.api_key }}"}`), matching legacy
`open-data-dc.go`'s `connsdk.APIKeyQuery("apikey", key)` exactly. The key is never logged.

## Streams notes

The MAR API is not paginated — each stream returns its full result set in a single response
(`pagination: {"type": "none"}`), matching legacy exactly.

- `locations` requests `GET /locations/{{ config.location }}` (the search term is embedded in the
  path, urlencoded by default per the engine's path-interpolation rule); `location` is required for
  this stream only (an unresolved `config.location` reference hard-errors, matching legacy's own
  explicit `"open-data-dc locations stream requires config location"` error). Records live at
  `Result.addresses`; each raw item is shaped `{"address":{"properties":{...}}, "distance": <num>}` —
  since the useful fields are nested two levels deep under `address.properties` (not at the record's
  top level), every field is extracted via `computed_fields` bare `{{ record.address.properties.<field>
  }}` references (typed extraction preserves `Latitude`/`Longitude`/`Xcoord`/`Ycoord` as numbers), plus
  a top-level `{{ record.distance }}` passthrough for the optional search-ranking score, matching
  legacy's `mapLocationRecord`/`locationProperties` fallback-lifting behavior field-for-field.
- `units` requests `GET /units/{{ config.marid }}`; `marid` is required for this stream only, matching
  legacy's explicit error. Records live at `Result.units` and are already flat (no nested wrap), so
  plain schema projection (no `computed_fields`) reproduces legacy's `mapUnitRecord` exactly.
- `ssls` requests `GET /ssls` with `marid` as an OPTIONAL query parameter
  (`"query": {"marid": {"template": "{{ config.marid }}", "omit_when_absent": true}}` — the engine's
  opt-in optional-query dialect), matching legacy's behavior of only setting the `marid` query param
  when configured, and reading every record otherwise. Records live at `Result.ssls` and are already
  flat, matching legacy's `mapSslRecord` exactly.

No stream has an incremental cursor: the MAR API is a read-only address-lookup service with no
modification timestamp, matching legacy's `CursorFields: nil` (full-refresh-only) exactly.

## Write actions & risks

None. Open Data DC's MAR 2 API is public and read-only; `capabilities.write` is `false` and this
bundle ships no `writes.json`.

## Known limits

- No stream supports incremental sync; every read is a full refresh, matching legacy exactly.
- `location` and `marid` are stream-scoped required config (only for `locations` and `units`
  respectively) rather than globally required `spec.json` properties, since `ssls` needs neither and
  legacy itself validates this per-stream, not globally — this mirrors legacy's exact per-stream
  validation in `requestFor`.
