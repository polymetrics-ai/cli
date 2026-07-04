# Overview

Breezometer is a wave2 fan-out declarative-HTTP migration, expanded in Pass B to the connector's
full documented `api.breezometer.com` surface. It reads BreezoMeter (now part of Google Maps
Platform's Environment APIs) point-in-time and forecast/history environmental data — air quality,
pollen, weather, and wildfire tracking — for a single configured lat/lon location, and writes a
stateless cleanest-route environmental-cleanliness scoring computation, through the BreezoMeter
REST API. The legacy package (`internal/connectors/breezometer`) stays registered and unchanged
until wave6's registry flip.

**A correction from the wave2-era research**: `docs.breezometer.com`'s live site now 301-redirects
to a generic Google Maps Platform marketing page (BreezoMeter's core Air Quality/Pollen data model
was absorbed into Google's own Air Quality/Pollen APIs), but the ORIGINAL per-product documentation
pages are still retrievable via Wayback Machine snapshots, and `api.breezometer.com` itself remains
a live, separately-branded endpoint distinct from Google's newer `airquality.googleapis.com`/
`pollen.googleapis.com`. This pass re-confirmed the real documented surface against those archived
pages rather than treating "the live doc site is gone" as license to leave the wave2-era
placeholder `excluded` entries (`historical/daily`, `road-segments`) unverified.

## Auth setup

Provide a BreezoMeter API key via the `api_key` secret; it is sent as the `key` query parameter
(`auth: api_key_query`) and never logged. `latitude`/`longitude` are required config values naming
the location every stream reads conditions/forecasts for.

## Streams notes

- `air_quality_current` — `GET /air-quality/v2/current-conditions`, a single-object response
  (`records.path: data`).
- `air_quality_forecast` — `GET /air-quality/v2/forecast/hourly`, a list response paginated via
  BreezoMeter's `next_page_token` body field (`pagination.type: cursor`, `token_path:
  next_page_token`). An optional `hours` query param (config `hours_to_forecast`) is sent only
  when configured (`omit_when_absent`).
- `air_quality_history` — `GET /air-quality/v2/historical/hourly`, same shape as forecast, with an
  optional `hours` query param sourced from config `historic_hours`. **This is the complete Air
  Quality v2 surface** — re-confirmed against the archived docs nav (Current Conditions/Hourly
  Forecast/Hourly History, exactly 3 sections, no daily/road-segment variant documented anywhere);
  the wave2-era `air-quality/v2/historical/daily` and `air-quality/v2/road-segments` `api_surface.json`
  placeholder entries did not correspond to any real endpoint and have been removed rather than
  carried forward.
- `pollen_forecast` — `GET /pollen/v2/forecast/daily`, same list/pagination shape, with an optional
  `days` query param sourced from config `days_to_forecast`. This is Pollen v2's ENTIRE documented
  surface (exactly one section: Daily Forecast).
- `weather_current` — `GET /weather/v1/current-conditions`, a single-object response.
- `weather_daily_forecast` **(new this pass)** — `GET /weather/v1/forecast/daily`, a previously
  unmodeled 3rd Weather v1 endpoint (the archived docs nav shows Current Conditions/Hourly
  Forecast/Daily Forecast as 3 sibling sections; only the first two were migrated in wave2). Records
  at `data`; cursor field `start_date` (each record's day-start timestamp); optional `days` query
  param (config `weather_days_to_forecast`, 1-5 per BreezoMeter's documented range).
- `wildfire_active_tracking` **(new this pass)** — `GET /fires/v1/locate-and-track` (Wildfire
  Tracker+'s "Precision Tracking / Active Fire" product), a self-serve `lat`/`lon`/`radius`-scoped
  endpoint returning active wildfires within the given radius. Records at `data`; primary key
  `EventId` (BreezoMeter's own globally-unique wildfire event id); cursor field `LastUpdated`.
  Requires the new `wildfire_radius_km` config value (BreezoMeter's documented range: 5-100 km).
  Field names are PascalCase (`EventId`/`CurrentLat`/`CalculatedAcres`/etc.) matching BreezoMeter's
  own documented wire shape for this product verbatim — a different casing convention than every
  other stream in this bundle, which is a real, confirmed BreezoMeter API inconsistency, not a
  documentation error on this bundle's part.
- `wildfire_burnt_area` **(new this pass)** — `GET /fires/v1/burnt-area` (Wildfire Tracker+'s
  "Precision Tracking / Burnt Area" product), the `lat`/`lon`/`radius`-scoped counterpart returning
  burnt-out former fire areas within a lookback window. Records at `data`; cursor field
  `ExtinguishedTS`. Optional `daysFromExtinguish` query param (config
  `wildfire_days_from_extinguish`) controls how long after a fire is marked extinguished its burnt
  area still appears.

Every location-scoped stream stamps the configured `latitude`/`longitude` onto every emitted record
via `computed_fields` (`{{ config.latitude }}`/`{{ config.longitude }}`).

Pagination: BreezoMeter's `next_page_token` cursor is read from the response body
(`pagination.type: cursor`, `cursor_param: page_token`, `token_path: next_page_token`) and requested
on the next page; pagination stops naturally when the token is absent/empty — the single-object
streams (`air_quality_current`, `weather_current`) and the wildfire/daily-weather streams never
carry a `next_page_token` in their real responses, so they always stop after one request.

Neither `page_size` nor `max_pages` are declared as `spec.json` config: the `cursor`/`token_path`
paginator has no page-size query param at all and no config-driven max-page override exists
anywhere in the engine's read path — declaring either as spec config would be dead, unwireable
config (F6, REVIEW.md).

## Write actions & risks

- `score_cleanest_route` — `POST /insights/v1/cleanest-route` (the Cleanest Route API): a
  **stateless** environmental-cleanliness scoring computation over caller-supplied route polylines.
  Unlike every other write action pattern in this migration, this is not a create/update/delete
  against a persistent BreezoMeter object — it computes and returns a Route Air Cleanliness Score
  (RACS, 0-100) per submitted route, with no lasting side effect on any BreezoMeter-held data.
  `kind: custom` reflects this: there is nothing to later read back or delete. Low risk; no approval
  required.

## Known limits

- **`pollen_forecast`'s `datetime` cursor field is computed from the raw API's `date` field, not a
  fallback chain.** Legacy's `pollenRecord` reads `item["datetime"]` and falls back to
  `item["date"]` only when `datetime` is absent (`breezometer/streams.go`'s `pollenRecord`); the
  daily pollen forecast endpoint's real wire shape has no `datetime` field at all (only `date`), so
  the fallback branch is legacy's defensive dead code for this stream, never actually exercised
  against the real API. The engine's `computed_fields` dialect has no conditional/fallback
  reference syntax, so this bundle maps `datetime` directly from `{{ record.date }}` — the same
  effective value legacy actually emits for every real pollen response.
- **Wildfire Tracker+'s Area Monitoring sub-product is not modeled** (`/fires/v1/area-monitoring/
  active` and `/fires/v1/area-monitoring/burnt`): its areas of interest are predefined at the
  CONTRACT level (the API key itself is scoped to a fixed set of areas configured by a BreezoMeter
  account manager) with no lat/lon/radius runtime parameter — not an ENGINE_GAP, just a capability
  this bundle's config surface cannot address without a contract-specific, out-of-band area
  definition. Precision Tracking's `locate-and-track`/`burnt-area` endpoints (both self-serve,
  lat/lon/radius-scoped like every other stream) cover the equivalent capability for a
  caller-chosen point instead.
- **Environmental Alerts (webhook-registration platform) and Heatmap Tile Overlay (raw PNG tiles,
  served from a distinct `tiles.breezometer.com` host) are out of scope**: neither fits this
  dialect's synchronous JSON request/response model — see `api_surface.json`'s `excluded` entries
  for the specific reasoning.
- `air_quality_current`/`weather_current`/`weather_daily_forecast`/`wildfire_active_tracking`/
  `wildfire_burnt_area` ship single-page fixtures (their real responses never paginate);
  `air_quality_forecast` ships the required 2-page fixture proving the `next_page_token` cursor
  advances and terminates.
