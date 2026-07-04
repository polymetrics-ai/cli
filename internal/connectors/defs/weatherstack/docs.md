# Overview

Weatherstack is a read-only declarative-HTTP connector for the Weatherstack REST API. It reads
current, historical, forecast, marine, and location-autocomplete weather data. This bundle was
originally migrated from `internal/connectors/weatherstack` (legacy wave2 fan-out:
`current`/`historical`/`forecast` only) and has since been expanded to the full documented
Weatherstack surface (Pass B). The legacy package stays registered and unchanged until wave6's
registry flip.

## Auth setup

Provide a Weatherstack API access key via the `access_key` secret; it is sent as the `access_key`
query parameter on every request (`auth: [{"mode": "api_key_query", "param": "access_key", ...}]`)
and is never logged. `base_url` defaults to `https://api.weatherstack.com` and may be overridden
for tests or proxies.

## Streams notes

5 streams: `current` (`GET /current`), `historical` (`GET /historical`), `forecast` (`GET
/forecast`), `marine` (`GET /marine`), `autocomplete` (`GET /autocomplete`).

`current`/`historical`/`forecast`/`marine` each return a single weather-report object (not an
array); `records.path: "."` extracts it as a one-element record set, matching legacy's
`recordsPath: "."` behavior (`connsdk.RecordsAt`'s single-object case). `autocomplete` returns a
genuine array under `results`.

`current`/`historical`/`forecast` send the `query` config value (location: city name, coordinates,
IP, or zip, including Weatherstack's documented semicolon-separated "bulk" multi-location shape —
a caller-supplied value shape, not a distinct connector feature) as the `query` query parameter.
`historical` additionally sends `historical_date` when set; `forecast` additionally sends
`forecast_days` when set — both matching legacy's conditional
`if value := ...; value != "" { q.Set(...) }` checks. All 4 weather-data streams
(`current`/`historical`/`forecast`/`marine`) now also forward optional `units`
(`m`=metric/`s`=scientific/`f`=Fahrenheit) and `language` (2-letter ISO code) query parameters when
configured, matching Weatherstack's own documented optional-parameter set common to every weather
endpoint.

`marine` (new) requires `latitude`/`longitude` config values instead of `query` — Weatherstack's
marine endpoint is coordinate-based, not location-string-based, per its own docs. `autocomplete`
(new) requires a separate `autocomplete_query` config value (kept independent of the `query` config
key the weather-data streams use, since a caller commonly wants a different, partial text value for
typeahead matching than the exact location string used for a weather-data read).

All streams declare `"projection": "passthrough"`: legacy's `Read` emits every decoded record
verbatim (`emit(connectors.Record(item))` in `internal/connectors/weatherstack/weatherstack.go`,
with no field-built `connectors.Record{...}` mapping — the only in-code touch is conditionally
back-filling `item["id"]` when absent, never dropping or renaming any other key). Schema-mode
projection would silently drop any Weatherstack response field not enumerated in
`schemas/{current,historical,forecast,marine,autocomplete}.json`, which is a meta-rule violation per
conventions.md §8 rule 1. The schemas remain a documentation surface listing the known/stable
fields; the live response may carry additional Weatherstack fields not enumerated there, and
passthrough mode ensures those still reach the record instead of being silently projected away.

The raw Weatherstack API response for `current`/`historical`/`forecast`/`marine` carries no `id`
field of its own. Legacy synthesizes one only when absent (`item["id"] == nil`) as
`"{stream}:{query}"` — in practice always, since the raw response never sets it. This bundle's
`computed_fields` unconditionally stamps the identical `"{stream}:{{ config.query }}"` value on
every `current`/`historical`/`forecast` record (and the analogous `"marine:{{ config.latitude
}},{{ config.longitude }}"` for `marine`), which reproduces legacy's actual runtime behavior on
every real API response byte-for-byte (the `computed_fields` dialect has no "only-if-absent"
conditional, but legacy's own guard is never actually false against the real API). `autocomplete`
has no legacy precedent (this is a new stream) and no single documented unique-id field of its own,
so its schema declares a composite `x-primary-key` (`name`/`region`/`country`/`lat`/`lon`) rather
than a synthesized id.

Because the id computation requires `config.query` to be resolvable at all times (`computed_fields`
templating hard-errors on an absent `config.*` key, unlike the opt-in-tolerant `stream.Query`
dialect), `query` is declared `required` in `spec.json` — a narrowing from legacy's optional
`query` (which would send an unfiltered request and likely receive a Weatherstack API error
response anyway, since Weatherstack itself requires a location query). See Known limits.

## Write actions & risks

None. Weatherstack is a read-only weather-data API with no mutation endpoints in legacy or in its
real documented surface; `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- `query` is `required` in `spec.json`, narrowing legacy's optional `query` config key. ACCEPTABLE
  per the parity-deviation meta-rule: legacy's own `queryParams` only conditionally sets `query`
  when non-empty, but every stream's synthesized `id` computed field needs `config.query` to
  resolve, and Weatherstack's real API requires a location parameter to return anything meaningful
  regardless — no accepted-input behavior that would produce a genuinely useful response is
  narrowed.
- No stream declares pagination: every Weatherstack endpoint covered here returns either a single
  report object or (for `autocomplete`) a small bounded results array with no documented pagination
  mechanism; this bundle issues exactly one request per stream, matching legacy's own single-request
  behavior for the 3 pre-existing streams and Weatherstack's own API shape for the 2 new ones.
- Weatherstack's "Bulk Queries" capability (semicolon-separated multiple locations in a single
  `query` value, Professional-plan+) is not modeled as a separate stream/feature — it is already
  expressible by a caller setting `config.query` to a semicolon-joined string, since this bundle
  forwards `query` verbatim.
