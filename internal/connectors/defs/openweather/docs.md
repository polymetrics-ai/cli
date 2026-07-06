# Overview

Reads current weather, hourly and daily forecasts, and government alerts for a configured geographic
location from the OpenWeather One Call API 3.0.

Readable streams: `current`, `hourly`, `daily`, `alerts`.

This connector is read-only; no write actions are declared.

Service API documentation: https://openweathermap.org/api.

## Auth setup

Connection fields:

- `appid` (required, secret, string); OpenWeather API key, sent as the appid query parameter; never
  logged.
- `base_url` (optional, string); default `https://api.openweathermap.org/data/3.0`; format `uri`;
  OpenWeather API base URL override for tests or proxies.
- `lang` (optional, string); Optional two-letter language code for weather condition text. Absent
  when unset.
- `lat` (required, string); Latitude of the location to read (e.g. "33.44").
- `lon` (required, string); Longitude of the location to read (e.g. "-94.04").
- `mode` (optional, string).
- `units` (optional, string); Optional unit system: standard, metric, or imperial. Absent when unset
  (OpenWeather's own default, standard/Kelvin, applies).

Secret fields are redacted in logs and write previews: `appid`.

Default configuration values: `base_url=https://api.openweathermap.org/data/3.0`.

Authentication behavior:

- API key authentication in query parameter `appid` using `secrets.appid`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/onecall?lat={{ config.lat }}&lon={{ config.lon
}}&exclude=minutely,hourly,daily,alerts`.

## Streams notes

Default pagination: single request; no pagination.

- `current`: GET `/onecall` - records path `current`; query `lang` from template `{{ config.lang
  }}`, omitted when absent; `lat`=`{{ config.lat }}`; `lon`=`{{ config.lon }}`; `units` from
  template `{{ config.units }}`, omitted when absent; computed output fields `lat`, `lon`;
  response-level fields copied to records `timezone`.
- `hourly`: GET `/onecall` - records path `hourly`; query `lang` from template `{{ config.lang }}`,
  omitted when absent; `lat`=`{{ config.lat }}`; `lon`=`{{ config.lon }}`; `units` from template `{{
  config.units }}`, omitted when absent; computed output fields `lat`, `lon`; response-level fields
  copied to records `timezone`.
- `daily`: GET `/onecall` - records path `daily`; query `lang` from template `{{ config.lang }}`,
  omitted when absent; `lat`=`{{ config.lat }}`; `lon`=`{{ config.lon }}`; `units` from template `{{
  config.units }}`, omitted when absent; computed output fields `lat`, `lon`, `temp_day`,
  `temp_max`, `temp_min`; response-level fields copied to records `timezone`.
- `alerts`: GET `/onecall` - records path `alerts`; query `lang` from template `{{ config.lang }}`,
  omitted when absent; `lat`=`{{ config.lat }}`; `lon`=`{{ config.lon }}`; `units` from template `{{
  config.units }}`, omitted when absent; computed output fields `lat`, `lon`; response-level fields
  copied to records `timezone`.

## Write actions & risks

This connector is read-only. Read behavior: external OpenWeather API read of public weather forecast
data.

## Known limits

- Batch defaults: read_page_size=1.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5.
