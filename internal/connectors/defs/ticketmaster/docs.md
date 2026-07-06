# Overview

Reads events, venues, attractions, and classifications from the Ticketmaster Discovery API.

Readable streams: `events`, `venues`, `attractions`, `classifications`.

This connector is read-only; no write actions are declared.

Service API documentation:
https://developer.ticketmaster.com/products-and-docs/apis/discovery-api/v2/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Ticketmaster Discovery API key, sent as the apikey query
  parameter. Never logged.
- `base_url` (optional, string); default `https://app.ticketmaster.com/discovery/v2`; format `uri`;
  Ticketmaster Discovery API base URL override for tests or proxies.
- `country_code` (optional, string); Optional ISO country code filter applied to the events stream.
- `keyword` (optional, string); Optional keyword search filter applied to the events stream.
- `locale` (optional, string); Optional locale filter applied to the events stream.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://app.ticketmaster.com/discovery/v2`.

Authentication behavior:

- API key authentication in query parameter `apikey` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/classifications.json`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `size`; starts at
0; page size 200.

- `events`: GET `/events.json` - records path `_embedded.events`; query `countryCode` from template
  `{{ config.country_code }}`, omitted when absent; `keyword` from template `{{ config.keyword }}`,
  omitted when absent; `locale` from template `{{ config.locale }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `size`; starts at 0; page size 200;
  emits passthrough records.
- `venues`: GET `/venues.json` - records path `_embedded.venues`; query `countryCode` from template
  `{{ config.country_code }}`, omitted when absent; `keyword` from template `{{ config.keyword }}`,
  omitted when absent; `locale` from template `{{ config.locale }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `size`; starts at 0; page size 200;
  emits passthrough records.
- `attractions`: GET `/attractions.json` - records path `_embedded.attractions`; query `countryCode`
  from template `{{ config.country_code }}`, omitted when absent; `keyword` from template `{{
  config.keyword }}`, omitted when absent; `locale` from template `{{ config.locale }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `size`; starts at 0;
  page size 200; emits passthrough records.
- `classifications`: GET `/classifications.json` - records path `_embedded.classifications`; query
  `countryCode` from template `{{ config.country_code }}`, omitted when absent; `keyword` from
  template `{{ config.keyword }}`, omitted when absent; `locale` from template `{{ config.locale
  }}`, omitted when absent; page-number pagination; page parameter `page`; size parameter `size`;
  starts at 0; page size 200; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Ticketmaster Discovery API read of public
event/venue data.

## Known limits

- Batch defaults: read_page_size=200.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=2.
