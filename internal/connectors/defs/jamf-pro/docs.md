# Overview

Reads Jamf Pro buildings, departments, categories, and scripts through the Jamf Pro REST API using
Basic-credential token-exchange authentication.

Readable streams: `buildings`, `departments`, `categories`, `scripts`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.jamf.com/jamf-pro/reference/classic-api.

## Auth setup

Connection fields:

- `base_url` (required, string); format `uri`; Jamf Pro API base URL, e.g.
  https://<subdomain>.jamfcloud.com/api. Required (not derived from a bare subdomain) - see docs.md
  Known limits.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-2000).
- `password` (required, secret, string); Jamf Pro API password, sent as the HTTP Basic password on
  the token exchange (POST /v1/auth/token). Never logged.
- `username` (required, string); Jamf Pro API username, sent as the HTTP Basic username on the token
  exchange.

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `max_pages=0`, `page_size=100`.

Authentication behavior:

- Connector-specific authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/buildings` with query `page`=`0`; `page-size`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `page-size`;
starts at 0; page size 100.

- `buildings`: GET `/v1/buildings` - records path `results`; page-number pagination; page parameter
  `page`; size parameter `page-size`; starts at 0; page size 100.
- `departments`: GET `/v1/departments` - records path `results`; page-number pagination; page
  parameter `page`; size parameter `page-size`; starts at 0; page size 100.
- `categories`: GET `/v1/categories` - records path `results`; page-number pagination; page
  parameter `page`; size parameter `page-size`; starts at 0; page size 100.
- `scripts`: GET `/v1/scripts` - records path `results`; page-number pagination; page parameter
  `page`; size parameter `page-size`; starts at 0; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external Jamf Pro API read of MDM configuration data
(buildings, departments, categories, scripts).

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1.
