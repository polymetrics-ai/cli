# Overview

Reads latest, currency-conversion, time-series, and fluctuation foreign-exchange rate data from the
exchangeratesapi.io REST API.

Readable streams: `latest`, `convert`, `timeseries`, `fluctuation`.

This connector is read-only; no write actions are declared.

Service API documentation: https://exchangeratesapi.io/documentation/.

## Auth setup

Connection fields:

- `access_key` (required, secret, string); exchangeratesapi.io access key, sent as the access_key
  query parameter. Never logged.
- `base` (optional, string); Optional base currency code applied to the latest and symbols requests
  (the API's optional 'base' query parameter).
- `base_url` (optional, string); default `https://api.exchangeratesapi.io/v1`; format `uri`;
  Exchange Rates API base URL override for tests or proxies.
- `convert_amount` (optional, string); Amount to convert for the convert stream (the API's required
  'amount' query parameter, sent verbatim as a string). Only read when the convert stream itself is
  synced.
- `convert_date` (optional, string); format `date`; Optional historical date (YYYY-MM-DD) for the
  convert stream (the API's optional 'date' query parameter); omitted for a live-rate conversion
  when unset.
- `convert_from` (optional, string); Source currency code for the convert stream (the API's required
  'from' query parameter). Only read when the convert stream itself is synced.
- `convert_to` (optional, string); Target currency code for the convert stream (the API's required
  'to' query parameter). Only read when the convert stream itself is synced.
- `fluctuation_end_date` (optional, string); format `date`; Inclusive window end date (YYYY-MM-DD)
  for the fluctuation stream (the API's required 'end_date' query parameter). Only read when the
  fluctuation stream itself is synced.
- `fluctuation_start_date` (optional, string); format `date`; Inclusive window start date
  (YYYY-MM-DD) for the fluctuation stream (the API's required 'start_date' query parameter). Only
  read when the fluctuation stream itself is synced.
- `mode` (optional, string).
- `timeseries_end_date` (optional, string); format `date`; Inclusive window end date (YYYY-MM-DD)
  for the timeseries stream (the API's required 'end_date' query parameter). Only read when the
  timeseries stream itself is synced.
- `timeseries_start_date` (optional, string); format `date`; Inclusive window start date
  (YYYY-MM-DD) for the timeseries stream (the API's required 'start_date' query parameter). Only
  read when the timeseries stream itself is synced.

Secret fields are redacted in logs and write previews: `access_key`.

Default configuration values: `base_url=https://api.exchangeratesapi.io/v1`.

Authentication behavior:

- API key authentication in query parameter `access_key` using `secrets.access_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/symbols`.

## Streams notes

Default pagination: single request; no pagination.

- `latest`: GET `/latest` - single-object response; records path `.`; query `base` from template `{{
  config.base }}`, omitted when absent.
- `convert`: GET `/convert` - single-object response; records path `.`; query `amount`=`{{
  config.convert_amount }}`; `date` from template `{{ config.convert_date }}`, omitted when absent;
  `from`=`{{ config.convert_from }}`; `to`=`{{ config.convert_to }}`.
- `timeseries`: GET `/timeseries` - single-object response; records path `.`; query `base` from
  template `{{ config.base }}`, omitted when absent; `end_date`=`{{ config.timeseries_end_date }}`;
  `start_date`=`{{ config.timeseries_start_date }}`.
- `fluctuation`: GET `/fluctuation` - single-object response; records path `.`; query `base` from
  template `{{ config.base }}`, omitted when absent; `end_date`=`{{ config.fluctuation_end_date }}`;
  `start_date`=`{{ config.fluctuation_start_date }}`.

## Write actions & risks

This connector is read-only. Read behavior: external exchangeratesapi.io read of public
foreign-exchange rate data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=2.
