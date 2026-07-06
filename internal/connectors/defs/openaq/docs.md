# Overview

Reads OpenAQ air quality reference data (countries, parameters, locations, instruments, and
manufacturers) from the OpenAQ v3 REST API.

Readable streams: `countries`, `parameters`, `locations`, `instruments`, `manufacturers`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.openaq.org/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); OpenAQ v3 API key, sent as the X-API-Key header; never
  logged.
- `base_url` (optional, string); default `https://api.openaq.org/v3`; format `uri`; OpenAQ API base
  URL override for tests or proxies.
- `countries_id` (optional, string); Optional comma-separated OpenAQ country id filter, applied to
  every stream via the countries_id query parameter. Absent when unset (no filter).
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.openaq.org/v3`.

Authentication behavior:

- API key authentication in `X-API-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/countries` with query `limit`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100.

- `countries`: GET `/countries` - records path `results`; query `countries_id` from template `{{
  config.countries_id }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100.
- `parameters`: GET `/parameters` - records path `results`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100.
- `locations`: GET `/locations` - records path `results`; query `countries_id` from template `{{
  config.countries_id }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100.
- `instruments`: GET `/instruments` - records path `results`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100.
- `manufacturers`: GET `/manufacturers` - records path `results`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external OpenAQ API read of public air-quality reference
data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=4.
