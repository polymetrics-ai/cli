# Overview

Reads NASA Open API data: Astronomy Picture of the Day, Near-Earth Objects (NeoWs feed and browse),
EPIC Earth imagery, and Mars rover photos. Read-only.

Readable streams: `apod`, `neo_feed`, `neo_browse`, `epic`, `mars_photos`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api.nasa.gov/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); NASA Open API key, sent as the api_key query parameter.
  Never logged.
- `base_url` (optional, string); default `https://api.nasa.gov`; format `uri`; NASA API base URL
  override for tests or proxies.
- `count` (optional, string); Number of random apod entries to return (apod only).
- `end_date` (optional, string); End date (YYYY-MM-DD) filter for apod and neo_feed.
- `mode` (optional, string).
- `sol` (optional, string); default `1000`; Martian sol (day) to filter mars_photos by.
- `start_date` (optional, string); Start date (YYYY-MM-DD) filter for apod and neo_feed.
- `thumbs` (optional, string); Set to true to include video thumbnail URLs (apod only).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.nasa.gov`, `sol=1000`.

Authentication behavior:

- API key authentication in query parameter `api_key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/planetary/apod`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `apod`, `neo_feed`, `epic`, `mars_photos`; page_number: `neo_browse`.

- `apod`: GET `/planetary/apod` - records at response root; query `count` from template `{{
  config.count }}`, omitted when absent; `end_date` from template `{{ config.end_date }}`, omitted
  when absent; `start_date` from template `{{ config.start_date }}`, omitted when absent; `thumbs`
  from template `{{ config.thumbs }}`, omitted when absent.
- `neo_feed`: GET `/neo/rest/v1/feed` - records path `near_earth_objects`; query `end_date` from
  template `{{ config.end_date }}`, omitted when absent; `start_date` from template `{{
  config.start_date }}`, omitted when absent.
- `neo_browse`: GET `/neo/rest/v1/neo/browse` - records path `near_earth_objects`; page-number
  pagination; page parameter `page`; no page-size parameter; starts at 0; page size 20; maximum 5
  page(s).
- `epic`: GET `/EPIC/api/natural` - records at response root.
- `mars_photos`: GET `/mars-photos/api/v1/rovers/curiosity/photos` - records path `photos`; query
  `sol` from template `{{ config.sol }}`, default `1000`; computed output fields `camera`, `rover`.

## Write actions & risks

This connector is read-only. Read behavior: external NASA Open API read of public astronomy and
space data.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=6.
