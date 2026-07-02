# Overview

Breezometer is a wave2 fan-out declarative-HTTP migration. It reads BreezoMeter (now part of
Google Maps Platform's Environment APIs) point-in-time and forecast/history environmental data —
air quality, pollen, and weather — for a single configured lat/lon location, through the
BreezoMeter REST API. This bundle targets full capability parity with
`internal/connectors/breezometer` (the hand-written connector it migrates); the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a BreezoMeter API key via the `api_key` secret; it is sent as the `key` query parameter
(`auth: api_key_query`) and never logged. `latitude`/`longitude` are required config values naming
the location every stream reads conditions/forecasts for.

## Streams notes

All 5 legacy streams are implemented:

- `air_quality_current` — `GET /air-quality/v2/current-conditions`, a single-object response
  (`records.path: data`, matching legacy's non-list `endpoint.list == false` shape).
- `air_quality_forecast` — `GET /air-quality/v2/forecast/hourly`, a list response paginated via
  BreezoMeter's `next_page_token` body field (`pagination.type: cursor`, `token_path:
  next_page_token`). An optional `hours` query param (config `hours_to_forecast`) is sent only
  when configured (`omit_when_absent`), matching legacy's conditional `base.Set("hours", hours)`.
- `air_quality_history` — `GET /air-quality/v2/historical/hourly`, same shape as forecast, with an
  optional `hours` query param sourced from config `historic_hours`.
- `pollen_forecast` — `GET /pollen/v2/forecast/daily`, same list/pagination shape, with an optional
  `days` query param sourced from config `days_to_forecast`.
- `weather_current` — `GET /weather/v1/current-conditions`, a single-object response.

Every stream stamps the configured `latitude`/`longitude` onto every emitted record via
`computed_fields` (`{{ config.latitude }}`/`{{ config.longitude }}`), matching legacy's
`record["latitude"] = lat; record["longitude"] = lng` injection in `harvest`. The composite primary
key `[datetime, latitude, longitude]` and cursor field `datetime` match legacy's stream catalog
across every stream.

Pagination: BreezoMeter's `next_page_token` cursor is read from the response body
(`pagination.type: cursor`, `cursor_param: page_token`, `token_path: next_page_token`) and requested
on the next page; pagination stops naturally when the token is absent/empty — the single-object
streams (`air_quality_current`, `weather_current`) never carry a `next_page_token` in their real
responses, so they always stop after one request, reproducing legacy's `!endpoint.list` early-stop
without needing a separate "is this stream a list" primitive. No `stop_path` is declared: legacy's
only stop signals are "single-object stream" (there is no server-side boolean equivalent) and
"empty token", both already covered.

Neither `page_size` nor `max_pages` are declared as `spec.json` config: the `cursor`/`token_path`
paginator has no page-size query param at all (BreezoMeter has no page-size control for these
endpoints) and no config-driven max-page override exists anywhere in the engine's read path
(`streams.json`'s `pagination` block is the only page-count/size lever the engine reads) —
declaring either as spec config would be dead, unwireable config (F6, REVIEW.md).

## Write actions & risks

None. BreezoMeter is a read-only environmental data source (`capabilities.write: false`); there is
no safe reverse-ETL write surface, matching legacy exactly.

## Known limits

- **`pollen_forecast`'s `datetime` cursor field is computed from the raw API's `date` field, not a
  fallback chain.** Legacy's `pollenRecord` reads `item["datetime"]` and falls back to
  `item["date"]` only when `datetime` is absent (`breezometer/streams.go`'s `pollenRecord`); the
  daily pollen forecast endpoint's real wire shape has no `datetime` field at all (only `date`), so
  the fallback branch is legacy's defensive dead code for this stream, never actually exercised
  against the real API. The engine's `computed_fields` dialect has no conditional/fallback
  reference syntax (a template is a fixed reference or filter chain, never an "A-or-B" expression),
  so this bundle maps `datetime` directly from `{{ record.date }}` — the same effective value legacy
  actually emits for every real pollen response. Documented here rather than left silent per the
  parity-deviation ledger meta-rule (conventions.md §5): this never changes emitted data for any
  real pollen API response, only for a hypothetical response shape legacy's own dead-code branch
  anticipated but the real API has never been observed to send. BreezoMeter's own API documentation
  is no longer reachable at its original URL (`docs.breezometer.com` now redirects to a generic
  Google Maps Platform marketing page, since BreezoMeter was folded into Google's Environment APIs)
  — legacy code is ground truth here per conventions.md's explicit precedence rule.
- Full BreezoMeter/Google Environment API surface (wildfire data, road-segment air quality, daily
  weather/history variants) is out of scope for this wave; see `api_surface.json`'s
  `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}` entries.
- `air_quality_current`/`weather_current` ship a single-page fixture (their real responses never
  paginate); `air_quality_forecast` ships the required 2-page fixture proving the
  `next_page_token` cursor advances and terminates.
