# Overview

Reads current, historical, forecast, marine, and location-autocomplete weather data from
Weatherstack. Read-only.

Readable streams: `current`, `historical`, `forecast`, `marine`, `autocomplete`.

This connector is read-only; no write actions are declared.

Service API documentation: https://weatherstack.com/documentation.

## Auth setup

Connection fields:

- `access_key` (required, secret, string); Weatherstack API access key. Sent as the access_key query
  parameter on every request; never logged.
- `autocomplete_query` (optional, string); Location text forwarded as the autocomplete stream's
  query parameter (partial city/region/zip/IP text for typeahead matching). Required for the
  autocomplete stream only; independent of the query key every weather-data stream uses.
- `base_url` (optional, string); default `https://api.weatherstack.com`; format `uri`; Weatherstack
  API base URL override for tests or proxies.
- `forecast_days` (optional, string); Number of days forwarded as the forecast_days query parameter
  on the forecast stream.
- `historical_date` (optional, string); Date (YYYY-MM-DD) forwarded as the historical_date query
  parameter on the historical stream.
- `language` (optional, string); Optional 2-letter ISO language code forwarded as the language query
  parameter on current/historical/forecast/marine when set.
- `latitude` (optional, string); Latitude forwarded as the marine stream's latitude query parameter.
  Required for the marine stream only (Weatherstack's marine endpoint takes coordinates, not the
  query location string every other stream uses).
- `longitude` (optional, string); Longitude forwarded as the marine stream's longitude query
  parameter. Required for the marine stream only.
- `mode` (optional, string).
- `query` (required, string); Location query (city name, coordinates, IP, or zip) forwarded as the
  query parameter on every stream's request.
- `units` (optional, string); Optional measurement unit system (m=metric, s=scientific,
  f=Fahrenheit) forwarded as the units query parameter on current/historical/forecast/marine when
  set. Weatherstack defaults to metric server-side when omitted.

Secret fields are redacted in logs and write previews: `access_key`.

Default configuration values: `base_url=https://api.weatherstack.com`.

Authentication behavior:

- API key authentication in query parameter `access_key` using `secrets.access_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/current`.

## Streams notes

Default pagination: single request; no pagination.

- `current`: GET `/current` - records path `.`; query `language` from template `{{ config.language
  }}`, omitted when absent; `query`=`{{ config.query }}`; `units` from template `{{ config.units
  }}`, omitted when absent; computed output fields `id`; emits passthrough records.
- `historical`: GET `/historical` - records path `.`; query `historical_date` from template `{{
  config.historical_date }}`, omitted when absent; `language` from template `{{ config.language }}`,
  omitted when absent; `query`=`{{ config.query }}`; `units` from template `{{ config.units }}`,
  omitted when absent; computed output fields `id`; emits passthrough records.
- `forecast`: GET `/forecast` - records path `.`; query `forecast_days` from template `{{
  config.forecast_days }}`, omitted when absent; `language` from template `{{ config.language }}`,
  omitted when absent; `query`=`{{ config.query }}`; `units` from template `{{ config.units }}`,
  omitted when absent; computed output fields `id`; emits passthrough records.
- `marine`: GET `/marine` - records path `.`; query `language` from template `{{ config.language
  }}`, omitted when absent; `latitude`=`{{ config.latitude }}`; `longitude`=`{{ config.longitude
  }}`; `units` from template `{{ config.units }}`, omitted when absent; computed output fields `id`;
  emits passthrough records.
- `autocomplete`: GET `/autocomplete` - records path `results`; query `query`=`{{
  config.autocomplete_query }}`; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Weatherstack API read of public weather data.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s).
