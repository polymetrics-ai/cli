# Overview

Reads Spotify Ads ad accounts, campaigns, ad sets, ads, businesses, business-scoped ad accounts, and
assets, and writes campaign mutations through the Spotify Ads API.

Readable streams: `ad_accounts`, `campaigns`, `ad_sets`, `ads`, `businesses`,
`business_ad_accounts`, `assets`.

Write actions: `update_campaign`.

Service API documentation: https://developer.spotify.com/documentation/ads-api.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Spotify Ads OAuth access token, sent as a Bearer token
  (Authorization: Bearer <access_token>). Never logged.
- `ad_account_id` (optional, string); Spotify Ads ad account ID; required by the campaigns, ad_sets,
  and ads streams to substitute the {ad_account_id} path segment.
- `base_url` (optional, string); default `https://api-partner.spotify.com/ads/v2`; format `uri`;
  Spotify Ads API base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api-partner.spotify.com/ads/v2`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/ad_accounts` with query `limit`=`1`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `businesses`, `business_ad_accounts`; offset_limit: `ad_accounts`,
`campaigns`, `ad_sets`, `ads`, `assets`.

- `ad_accounts`: GET `/ad_accounts` - records path `ad_accounts`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `campaigns`: GET `/ad_accounts/{{ config.ad_account_id }}/campaigns` - records path `campaigns`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `ad_sets`: GET `/ad_accounts/{{ config.ad_account_id }}/ad_sets` - records path `ad_sets`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `ads`: GET `/ad_accounts/{{ config.ad_account_id }}/ads` - records path `ads`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `businesses`: GET `/businesses` - records path `businesses`.
- `business_ad_accounts`: GET `/businesses/{{ fanout.id }}/ad_accounts` - records path
  `ad_accounts`; fan-out; ids from request `/businesses`; id-list records path `businesses`; id
  field `id`; id inserted into the request path; stamps `business_id`.
- `assets`: GET `/ad_accounts/{{ config.ad_account_id }}/assets` - records path `assets`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100.

## Write actions & risks

Overall write risk: external Spotify Ads API mutation of a campaign's name, purchase order
reference, or status (active/paused/archived).

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `update_campaign`: PATCH `/ad_accounts/{{ record.ad_account_id }}/campaigns/{{ record.id }}` -
  kind `update`; body type `json`; path fields `ad_account_id`, `id`; required record fields
  `ad_account_id`, `id`; accepted fields `ad_account_id`, `id`, `name`, `purchase_order`, `status`;
  risk: mutates a live campaign's name, purchase-order reference, or status; setting status to
  PAUSED/ARCHIVED stops that campaign's ad delivery and spend, approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 7 stream-backed endpoint group(s), 1 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=6, destructive_admin=5, duplicate_of=7, non_data_endpoint=4, out_of_scope=65,
  requires_elevated_scope=2.
