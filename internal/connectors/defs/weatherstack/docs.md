# Overview

Weatherstack is a read-only declarative-HTTP connector migrated from
`internal/connectors/weatherstack` (legacy wave2 fan-out). It reads current, historical, and
forecast weather data from the Weatherstack REST API. This bundle is capability-parity with the
legacy hand-written connector; the legacy package stays registered and unchanged until wave6's
registry flip.

## Auth setup

Provide a Weatherstack API access key via the `access_key` secret; it is sent as the `access_key`
query parameter on every request (`auth: [{"mode": "api_key_query", "param": "access_key", ...}]`)
and is never logged. `base_url` defaults to `https://api.weatherstack.com` and may be overridden
for tests or proxies.

## Streams notes

3 streams: `current` (`GET /current`), `historical` (`GET /historical`), `forecast` (`GET
/forecast`). Every stream response is a single weather-report object (not an array); `records.path:
"."` extracts it as a one-element record set, matching legacy's `recordsPath: "."` behavior
(`connsdk.RecordsAt`'s single-object case).

Every stream sends the `query` config value (location: city name, coordinates, IP, or zip) as the
`query` query parameter. `historical` additionally sends `historical_date` when set; `forecast`
additionally sends `forecast_days` when set — both matching legacy's conditional
`if value := ...; value != "" { q.Set(...) }` checks.

All 3 streams declare `"projection": "passthrough"`: legacy's `Read` emits every decoded record
verbatim (`emit(connectors.Record(item))` in `internal/connectors/weatherstack/weatherstack.go`,
with no field-built `connectors.Record{...}` mapping — the only in-code touch is conditionally
back-filling `item["id"]` when absent, never dropping or renaming any other key). Schema-mode
projection would silently drop any Weatherstack response field not enumerated in
`schemas/{current,historical,forecast}.json`, which is a meta-rule violation per conventions.md §8
rule 1. The schemas remain a documentation surface listing the known/stable fields; the live
response may carry additional Weatherstack fields not enumerated there, and passthrough mode
ensures those still reach the record instead of being silently projected away.

The raw Weatherstack API response carries no `id` field of its own. Legacy synthesizes one only
when absent (`item["id"] == nil`) as `"{stream}:{query}"` — in practice always, since the raw
response never sets it. This bundle's `computed_fields` unconditionally stamps the identical
`"{stream}:{{ config.query }}"` value on every record, which reproduces legacy's actual runtime
behavior on every real API response byte-for-byte (the `computed_fields` dialect has no
"only-if-absent" conditional, but legacy's own guard is never actually false against the real API).
Because the id computation requires `config.query` to be resolvable at all times (`computed_fields`
templating hard-errors on an absent `config.*` key, unlike the opt-in-tolerant `stream.Query`
dialect), `query` is declared `required` in `spec.json` — a narrowing from legacy's optional
`query` (which would send an unfiltered request and likely receive a Weatherstack API error
response anyway, since Weatherstack itself requires a location query). See Known limits.

## Write actions & risks

None. Weatherstack is a read-only weather-data API with no mutation endpoints in legacy;
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- Only the 3 legacy-parity read streams are implemented; other Weatherstack endpoints (marine
  weather, autocomplete, time-series-aggregate) are out of scope for this migration wave — see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}`
  entries.
- `query` is `required` in `spec.json`, narrowing legacy's optional `query` config key. ACCEPTABLE
  per the parity-deviation meta-rule: legacy's own `queryParams` only conditionally sets `query`
  when non-empty, but every stream's synthesized `id` computed field needs `config.query` to
  resolve, and Weatherstack's real API requires a location parameter to return anything meaningful
  regardless — no accepted-input behavior that would produce a genuinely useful response is
  narrowed.
- Neither stream declares pagination: legacy's `Read` issues exactly one request per stream with no
  paging loop, and this bundle mirrors that (no `pagination` block in `streams.json`).
