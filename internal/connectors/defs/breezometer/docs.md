# Overview

Reads BreezoMeter (Google Environment) air quality, pollen, weather, and wildfire-tracking
conditions/forecasts for a configured location via the BreezoMeter REST API; writes a stateless
cleanest-route environmental-cleanliness scoring computation.

Readable streams: `air_quality_current`, `air_quality_forecast`, `air_quality_history`,
`pollen_forecast`, `weather_current`, `weather_daily_forecast`, `wildfire_active_tracking`,
`wildfire_burnt_area`.

Write actions: `score_cleanest_route`.

Service API documentation: https://docs.breezometer.com/api-documentation/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); BreezoMeter API key, sent as the `key` query parameter;
  never logged.
- `base_url` (optional, string); default `https://api.breezometer.com`; format `uri`; BreezoMeter
  API base URL override for tests or proxies.
- `days_to_forecast` (optional, string); Optional `days` query param for the pollen_forecast stream
  (number of daily forecast points to return).
- `historic_hours` (optional, string); Optional `hours` query param for the air_quality_history
  stream (number of hourly historical points to return).
- `hours_to_forecast` (optional, string); Optional `hours` query param for the air_quality_forecast
  stream (number of hourly forecast points to return).
- `latitude` (required, string); Latitude of the location to read conditions/forecasts for.
- `longitude` (required, string); Longitude of the location to read conditions/forecasts for.
- `mode` (optional, string).
- `weather_days_to_forecast` (optional, string); Optional `days` query param for the
  weather_daily_forecast stream (number of daily weather forecast points to return, 1-5).
- `wildfire_days_from_extinguish` (optional, string); Optional `daysFromExtinguish` query param for
  the wildfire_burnt_area stream (days since a fire was marked extinguished to still include its
  burnt area).
- `wildfire_radius_km` (optional, string); Required `radius` query param (kilometers) for the
  wildfire_active_tracking and wildfire_burnt_area streams: desired search radius from
  latitude/longitude, 5-100 km.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.breezometer.com`.

Authentication behavior:

- API key authentication in query parameter `key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/air-quality/v2/current-conditions` with query `features`=`local_aqi`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page_token`; next token from
`next_page_token`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `air_quality_current`: GET `/air-quality/v2/current-conditions` - records path `data`; query
  `features`=`local_aqi`; `lat`=`{{ config.latitude }}`; `lon`=`{{ config.longitude }}`; cursor
  pagination; cursor parameter `page_token`; next token from `next_page_token`; incremental cursor
  `datetime`; formatted as `rfc3339`; computed output fields `latitude`, `longitude`.
- `air_quality_forecast`: GET `/air-quality/v2/forecast/hourly` - records path `data`; query
  `features`=`local_aqi`; `hours` from template `{{ config.hours_to_forecast }}`, omitted when
  absent; `lat`=`{{ config.latitude }}`; `lon`=`{{ config.longitude }}`; cursor pagination; cursor
  parameter `page_token`; next token from `next_page_token`; incremental cursor `datetime`;
  formatted as `rfc3339`; computed output fields `latitude`, `longitude`.
- `air_quality_history`: GET `/air-quality/v2/historical/hourly` - records path `data`; query
  `features`=`local_aqi`; `hours` from template `{{ config.historic_hours }}`, omitted when absent;
  `lat`=`{{ config.latitude }}`; `lon`=`{{ config.longitude }}`; cursor pagination; cursor parameter
  `page_token`; next token from `next_page_token`; incremental cursor `datetime`; formatted as
  `rfc3339`; computed output fields `latitude`, `longitude`.
- `pollen_forecast`: GET `/pollen/v2/forecast/daily` - records path `data`; query `days` from
  template `{{ config.days_to_forecast }}`, omitted when absent; `lat`=`{{ config.latitude }}`;
  `lon`=`{{ config.longitude }}`; cursor pagination; cursor parameter `page_token`; next token from
  `next_page_token`; incremental cursor `datetime`; formatted as `rfc3339`; computed output fields
  `datetime`, `latitude`, `longitude`.
- `weather_current`: GET `/weather/v1/current-conditions` - records path `data`; query `lat`=`{{
  config.latitude }}`; `lon`=`{{ config.longitude }}`; cursor pagination; cursor parameter
  `page_token`; next token from `next_page_token`; incremental cursor `datetime`; formatted as
  `rfc3339`; computed output fields `latitude`, `longitude`.
- `weather_daily_forecast`: GET `/weather/v1/forecast/daily` - records path `data`; query `days`
  from template `{{ config.weather_days_to_forecast }}`, omitted when absent; `lat`=`{{
  config.latitude }}`; `lon`=`{{ config.longitude }}`; cursor pagination; cursor parameter
  `page_token`; next token from `next_page_token`; incremental cursor `start_date`; formatted as
  `rfc3339`; computed output fields `latitude`, `longitude`.
- `wildfire_active_tracking`: GET `/fires/v1/locate-and-track` - records path `data`; query
  `lat`=`{{ config.latitude }}`; `lon`=`{{ config.longitude }}`; `radius`=`{{
  config.wildfire_radius_km }}`; cursor pagination; cursor parameter `page_token`; next token from
  `next_page_token`; computed output fields `latitude`, `longitude`.
- `wildfire_burnt_area`: GET `/fires/v1/burnt-area` - records path `data`; query
  `daysFromExtinguish` from template `{{ config.wildfire_days_from_extinguish }}`, omitted when
  absent; `lat`=`{{ config.latitude }}`; `lon`=`{{ config.longitude }}`; `radius`=`{{
  config.wildfire_radius_km }}`; cursor pagination; cursor parameter `page_token`; next token from
  `next_page_token`; computed output fields `latitude`, `longitude`.

## Write actions & risks

Overall write risk: stateless environmental-cleanliness scoring computation over caller-supplied
route geometries; no persistent BreezoMeter object is created or mutated.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `score_cleanest_route`: POST `/insights/v1/cleanest-route` - kind `custom`; body type `json`;
  required record fields `routes`; accepted fields `routes`; risk: stateless
  environmental-cleanliness scoring computation over caller-supplied route geometries; creates or
  mutates no persistent BreezoMeter object and has no side effects beyond the API call itself,
  low-risk.

## Known limits

- API coverage includes 8 stream-backed endpoint group(s), 1 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, requires_elevated_scope=3.
