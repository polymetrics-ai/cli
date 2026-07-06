# Overview

Reads Apple Search Ads campaigns, ad groups, targeting keywords, and ads via the Apple Search Ads
Campaign Management API using an OAuth2 client-credentials grant scoped to an organization.
Read-only.

Readable streams: `campaigns`, `adgroups`, `keywords`, `ads`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.apple.com/documentation/apple_search_ads.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.searchads.apple.com/api/v5`; format `uri`;
  Apple Search Ads Campaign Management API base URL.
- `client_id` (required, secret, string); Apple Search Ads API client ID, used as client_id in the
  OAuth2 client_credentials token exchange. Never logged.
- `client_secret` (required, secret, string); Apple Search Ads API client secret, used as
  client_secret in the OAuth2 client_credentials token exchange. Never logged.
- `max_pages` (optional, string); default `0`.
- `org_id` (required, string); Apple Search Ads organization ID.
- `page_size` (optional, string); default `1000`; Records per page (offset/limit pagination),
  1-1000.
- `token_refresh_endpoint` (optional, string); default
  `https://appleid.apple.com/auth/oauth2/token`; format `uri`; Apple ID OAuth2 token endpoint
  override.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`.

Default configuration values: `base_url=https://api.searchads.apple.com/api/v5`, `max_pages=0`,
`page_size=1000`, `token_refresh_endpoint=https://appleid.apple.com/auth/oauth2/token`.

Authentication behavior:

- OAuth 2.0 client credentials authentication using `config.token_refresh_endpoint`,
  `secrets.client_id`, `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `campaigns` with query `limit`=`1`; `offset`=`0`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 1000.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `campaigns`: GET `campaigns` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 1000; incremental cursor `modification_time`;
  formatted as `rfc3339`; computed output fields `ad_channel_type`, `billing_event`,
  `budget_amount`, `countries_or_regions`, `creation_time`, `daily_budget_amount`, `deleted`,
  `display_status`, `id`, `modification_time`, `name`, `org_id`, `serving_status`, `status`,
  `supply_sources`.
- `adgroups`: POST `adgroups/find` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 1000; incremental cursor `modification_time`;
  formatted as `rfc3339`; computed output fields `campaign_id`, `cpa_goal`, `creation_time`,
  `default_bid_amount`, `deleted`, `display_status`, `end_time`, `id`, `modification_time`, `name`,
  `pricing_model`, `serving_status`, `start_time`, `status`.
- `keywords`: POST `targetingkeywords/find` - records path `data`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 1000; incremental cursor
  `modification_time`; formatted as `rfc3339`; computed output fields `ad_group_id`, `bid_amount`,
  `campaign_id`, `deleted`, `id`, `match_type`, `modification_time`, `status`, `text`.
- `ads`: POST `ads/find` - records path `data`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 1000; incremental cursor `modification_time`; formatted as
  `rfc3339`; computed output fields `ad_group_id`, `campaign_id`, `creation_time`, `creative_id`,
  `creative_type`, `deleted`, `id`, `modification_time`, `name`, `serving_status`, `status`.

## Write actions & risks

This connector is read-only. Read behavior: external Apple Search Ads API read of campaign, ad
group, keyword, and ad data.

## Known limits

- Batch defaults: read_page_size=1000.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=4, requires_elevated_scope=1.
