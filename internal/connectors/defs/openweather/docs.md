# Overview

OpenWeather's One Call API 3.0 returns one JSON document per geographic coordinate pair, with
`current`/`hourly`/`daily`/`alerts` sections. This bundle reads all four sections as four streams
for a single configured location. It is migrated from `internal/connectors/openweather` (the
hand-written connector this bundle replaces at parity); the legacy package stays registered and
unchanged until wave6's registry flip.

## Auth setup

Provide an OpenWeather API key via the `appid` secret; it is sent as the `appid` query parameter
(`auth: [{"mode": "api_key_query", "param": "appid", ...}]`) and is never logged.

## Streams notes

Each stream (`current`, `hourly`, `daily`, `alerts`) issues a `GET /onecall?lat=...&lon=...`
request (no pagination — the One Call endpoint is not paginated) and extracts its section from the
single response document: `current` is a single object (`records.path: "current"`, one record),
`hourly`/`daily`/`alerts` are arrays. `lat`/`lon` are required config values stamped onto every
record via `computed_fields` so rows stay location-identifiable, matching legacy's `annotate`.
Optional `units` (`standard`/`metric`/`imperial`) and `lang` query parameters are wired via the
`omit_when_absent` optional-query dialect, matching legacy's conditional `query.Set`. `daily`'s
nested `temp: {day, min, max}` object is flattened into `temp_day`/`temp_min`/`temp_max` via
dotted-path `computed_fields` (`{{ record.temp.day }}` etc.), matching legacy's `mapDaily`.
`dt` (Unix seconds) is both the primary-key tiebreaker and incremental cursor for the three
time-series streams; `alerts` uses `start` as its cursor, matching legacy's `CursorFields`.

## Write actions & risks

None. OpenWeather is a read-only weather API; `capabilities.write` is `false` and no
`writes.json` is shipped, matching legacy's `ErrUnsupportedOperation` `Write` stub.

## Known limits

- **Multi-location fan-out is out of scope (documented scope narrowing).** Legacy accepted either
  a single `lat`/`lon` pair OR a semicolon-separated `locations` list (`"lat1,lon1;lat2,lon2"`),
  issuing one `/onecall` request per location and emitting records from every location in a single
  `Read` call. The engine's declarative read path issues exactly one request (or one paginated
  series of requests) per `Read` call with no config-driven multi-request fan-out primitive — this
  is exactly the "sub-resource fan-out reads" Tier-2 `StreamHook` trigger named in
  `docs/migration/conventions.md` §1's Tier-2 table. This bundle instead requires a single `lat`
  and `lon` (both `required` in `spec.json`); a caller previously configuring exactly one location
  (the common case, and legacy's own single-location code path) sees byte-identical behavior. A
  caller relying on the `locations` multi-value convenience must configure one connector instance
  per location instead (or a future Tier-2 hook wave can add a `StreamHook` for the fan-out without
  changing this bundle's single-location shape).
- **`weather[0]`-derived scalar fields (`weather_main`, `weather_description`, `weather_icon`) are
  dropped from parity (documented scope narrowing, not silently wrong).** Legacy's `addWeather`
  lifts the first element of the raw `weather` array into three scalar columns on every
  `current`/`hourly`/`daily` record. The engine's `computed_fields`/schema-projection dialect has
  no array-index dereference (`resolveRecordPathValue`/`selectPath` walk `map[string]any` only; a
  dotted path segment into a JSON array is unsupported, confirmed against
  `internal/connectors/engine/interpolate.go`) — there is no declarative way to reach
  `record.weather.0.main`. Each stream's schema instead retains the raw `weather` array field
  verbatim (schema-projection default), preserving 100% of the underlying data (nothing is lost,
  only the derived scalar convenience columns); a downstream consumer can still recover
  `weather[0].main`/`.description`/`.icon` from the array. This is an `ENGINE_GAP` candidate (no
  array-index primitive exists anywhere in the dialect) — if a future engine increment adds one,
  re-tighten these three fields back in.
- **The `timezone` sibling field is dropped from parity for the same reason.** Legacy reads the
  response's top-level `timezone` string once per request and stamps it onto every record of every
  section. `computed_fields` templates resolve only against the record extracted at
  `stream.records.path` (confirmed against `internal/connectors/engine/read.go`'s
  `applyComputedFields`/`rawRecords` scoping) — a sibling top-level field outside that extracted
  scope (like `timezone`, which lives beside `current`/`hourly`/`daily`/`alerts`, not inside any of
  them) is unreachable from any stream's computed_fields. Not modeled in this migration; the same
  `ENGINE_GAP` note applies.
- The wider OpenWeather product surface (5-day/3-hour forecast, air pollution, geocoding,
  historical weather via `/onecall/timemachine`) is out of scope for this wave; see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}`
  entries.
