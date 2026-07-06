# Overview

Reads TikTok Business advertisers, campaigns, ad groups, and ads through the TikTok Marketing
(Business) API.

Readable streams: `advertisers`, `campaigns`, `adgroups`, `ads`.

This connector is read-only; no write actions are declared.

Service API documentation: https://business-api.tiktok.com/portal/docs?id=1740029169927169.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); TikTok Business API access token, sent as the
  Access-Token header (not Bearer). Never logged.
- `advertiser_id` (optional, string); Optional TikTok advertiser id filter. Sent as advertiser_id on
  campaigns/adgroups/ads streams, and as a JSON-array advertiser_ids on the advertisers stream.
- `base_url` (optional, string); default `https://business-api.tiktok.com/open_api/v1.3`; format
  `uri`; TikTok Business API base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://business-api.tiktok.com/open_api/v1.3`.

Authentication behavior:

- API key authentication in `Access-Token` using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/advertiser/info/`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `page_size`;
starts at 1; page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `advertisers`: GET `/advertiser/info/` - records path `data.list`; query `advertiser_ids` from
  template `["{{ config.advertiser_id }}"]`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `page_size`; starts at 1; page size 100.
- `campaigns`: GET `/campaign/get/` - records path `data.list`; query `advertiser_id` from template
  `{{ config.advertiser_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `page_size`; starts at 1; page size 100; incremental cursor `modify_time`;
  formatted as `rfc3339`.
- `adgroups`: GET `/adgroup/get/` - records path `data.list`; query `advertiser_id` from template
  `{{ config.advertiser_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `page_size`; starts at 1; page size 100; incremental cursor `modify_time`;
  formatted as `rfc3339`.
- `ads`: GET `/ad/get/` - records path `data.list`; query `advertiser_id` from template `{{
  config.advertiser_id }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `page_size`; starts at 1; page size 100; incremental cursor `modify_time`; formatted as
  `rfc3339`.

## Write actions & risks

This connector is read-only. Read behavior: external TikTok Business API read of advertiser and
campaign management data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=2, out_of_scope=1.
