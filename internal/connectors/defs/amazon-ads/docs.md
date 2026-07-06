# Overview

Reads Amazon Advertising profiles, Sponsored Products campaigns, ad groups, product ads, keywords,
negative keywords, and portfolios via the Amazon Ads API using a Login with Amazon (LWA)
refresh-token grant. Read-only.

Readable streams: `profiles`, `campaigns`, `ad_groups`, `portfolios`, `keywords`, `product_ads`,
`negative_keywords`.

This connector is read-only; no write actions are declared.

Service API documentation: https://advertising.amazon.com/API/docs/en-us/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://advertising-api.amazon.com`; format `uri`; Amazon
  Ads API base URL. Set to https://advertising-api-eu.amazon.com or
  https://advertising-api-fe.amazon.com for EU/FE accounts.
- `client_id` (required, secret, string); Login with Amazon (LWA) application client ID, used both
  as client_id in the refresh_token token exchange and verbatim in the
  Amazon-Advertising-API-ClientId header on every data request. Never logged.
- `client_secret` (required, secret, string); Login with Amazon (LWA) application client secret,
  used as client_secret in the refresh_token token exchange. Never logged.
- `max_pages` (optional, string); default `0`.
- `page_size` (optional, string); default `100`; Records per page (startIndex/count offset
  pagination), 1-100.
- `profile_id` (optional, string); Amazon Ads advertising profile ID.
- `refresh_token` (required, secret, string); Long-lived LWA refresh_token exchanged for a
  short-lived access_token on every Check/Read. Never logged.
- `token_url` (optional, string); default `https://api.amazon.com/auth/o2/token`; format `uri`; Set
  to https://api.amazon.co.uk/auth/o2/token or https://api.amazon.co.jp/auth/o2/token for EU/FE
  accounts.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`,
`refresh_token`.

Default configuration values: `base_url=https://advertising-api.amazon.com`, `max_pages=0`,
`page_size=100`, `token_url=https://api.amazon.com/auth/o2/token`.

Authentication behavior:

- Connector-specific authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `v2/profiles` with query `count`=`1`; `startIndex`=`0`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `startIndex`; limit parameter `count`;
page size 100.

- `profiles`: GET `v2/profiles` - records at response root; offset/limit pagination; offset
  parameter `startIndex`; limit parameter `count`; page size 100; computed output fields
  `account_id`, `account_name`, `account_type`, `country_code`, `currency_code`, `daily_budget`,
  `marketplace_string_id`, `profile_id`, `timezone`.
- `campaigns`: GET `v2/sp/campaigns` - records at response root; offset/limit pagination; offset
  parameter `startIndex`; limit parameter `count`; page size 100; computed output fields
  `campaign_id`, `campaign_type`, `daily_budget`, `end_date`, `name`, `portfolio_id`,
  `premium_bid_adjustment`, `start_date`, `state`, `targeting_type`.
- `ad_groups`: GET `v2/sp/adGroups` - records at response root; offset/limit pagination; offset
  parameter `startIndex`; limit parameter `count`; page size 100; computed output fields
  `ad_group_id`, `campaign_id`, `default_bid`, `name`, `state`.
- `portfolios`: GET `v2/portfolios` - records at response root; offset/limit pagination; offset
  parameter `startIndex`; limit parameter `count`; page size 100; computed output fields
  `in_budget`, `name`, `portfolio_id`, `state`.
- `keywords`: GET `v2/sp/keywords` - records at response root; offset/limit pagination; offset
  parameter `startIndex`; limit parameter `count`; page size 100; computed output fields
  `ad_group_id`, `bid`, `campaign_id`, `keyword_id`, `keyword_text`, `match_type`, `state`.
- `product_ads`: GET `v2/sp/productAds` - records at response root; offset/limit pagination; offset
  parameter `startIndex`; limit parameter `count`; page size 100; computed output fields
  `ad_group_id`, `ad_id`, `asin`, `campaign_id`, `serving_status`, `sku`, `state`.
- `negative_keywords`: GET `v2/sp/negativeKeywords` - records at response root; offset/limit
  pagination; offset parameter `startIndex`; limit parameter `count`; page size 100; computed output
  fields `ad_group_id`, `campaign_id`, `keyword_id`, `keyword_text`, `match_type`, `state`.

## Write actions & risks

This connector is read-only. Read behavior: external Amazon Ads API read of profile, campaign, ad
group, product ad, keyword, negative keyword, and portfolio data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 7 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5.
