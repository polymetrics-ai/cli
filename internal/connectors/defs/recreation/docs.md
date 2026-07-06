# Overview

Reads Recreation.gov RIDB facilities, campsites, activities, organizations, and recreation areas
through the RIDB REST API.

Readable streams: `facilities`, `campsites`, `activities`, `organizations`, `recareas`.

This connector is read-only; no write actions are declared.

Service API documentation: https://ridb.recreation.gov/docs.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); RIDB API key, sent as the apikey header (apikey: <api_key>).
  Never logged.
- `base_url` (optional, string); default `https://ridb.recreation.gov/api/v1`; format `uri`;
  Recreation.gov RIDB API base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://ridb.recreation.gov/api/v1`.

Authentication behavior:

- API key authentication in `apikey` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/facilities` with query `limit`=`1`; `offset`=`0`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 50.

- `facilities`: GET `/facilities` - records path `RECDATA`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 50; computed output fields `id`, `name`,
  `type`, `updated_at`.
- `campsites`: GET `/campsites` - records path `RECDATA`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 50; computed output fields `id`, `name`, `type`,
  `updated_at`.
- `activities`: GET `/activities` - records path `RECDATA`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 50; computed output fields `id`, `name`.
- `organizations`: GET `/organizations` - records path `RECDATA`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 50; computed output fields `id`, `name`.
- `recareas`: GET `/recareas` - records path `RECDATA`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 50; computed output fields `id`, `name`,
  `updated_at`.

## Write actions & risks

This connector is read-only. Read behavior: external Recreation.gov RIDB API read of public
facility, campsite, activity, organization, and recreation-area data.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=7.
