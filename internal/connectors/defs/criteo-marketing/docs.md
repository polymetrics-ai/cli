# Overview

Reads Criteo Marketing Solutions ad sets, advertisers, campaigns, audiences, ad spend statistics,
and Marketplace Performance Outcomes (MPO) advertisers/sellers/budgets/seller-campaigns through the
Criteo REST API using OAuth2 client-credentials auth.

Readable streams: `ad_sets`, `advertisers`, `campaigns`, `audiences`, `statistics`,
`mpo_advertisers`, `mpo_sellers`, `mpo_budgets`, `mpo_seller_campaigns`.

This connector is read-only; no write actions are declared.

Service API documentation:
https://developers.criteo.com/marketing-solutions/reference/getting-started.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.criteo.com`; format `uri`; Criteo API base URL
  override for tests or proxies.
- `client_id` (required, secret, string); Criteo Marketing Solutions API client ID (OAuth2
  client-credentials). Never logged.
- `client_secret` (required, secret, string); Criteo Marketing Solutions API client secret (OAuth2
  client-credentials). Never logged.
- `currency` (optional, string); Statistics report currency filter (e.g. USD, EUR); also used to
  stamp the fixture-mode Currency field on other streams.
- `end_date` (optional, string); format `date`; Statistics report upper-bound date (YYYY-MM-DD),
  sent as the report's endDate filter.
- `start_date` (optional, string); format `date`; Statistics report lower-bound date (YYYY-MM-DD),
  sent as the report's startDate filter.
- `token_url` (optional, string); default `https://api.criteo.com/oauth2/token`; format `uri`.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`.

Default configuration values: `base_url=https://api.criteo.com`,
`token_url=https://api.criteo.com/oauth2/token`.

Authentication behavior:

- OAuth 2.0 client credentials authentication using `config.token_url`, `secrets.client_id`,
  `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/2024-01/marketing-solutions/advertisers`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

Pagination by stream: none: `mpo_advertisers`, `mpo_sellers`, `mpo_budgets`, `mpo_seller_campaigns`;
offset_limit: `ad_sets`, `advertisers`, `campaigns`, `audiences`, `statistics`.

- `ad_sets`: GET `/2024-01/marketing-solutions/ad-sets/search` - records path `data`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100; computed output
  fields `advertiserId`, `campaignId`, `datasetId`, `destinationEnvironment`, `mediaType`, `name`,
  `objective`, `status`.
- `advertisers`: GET `/2024-01/marketing-solutions/advertisers` - records path `data`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100; computed output
  fields `country`, `currency`, `name`, `timezone`.
- `campaigns`: GET `/2024-01/marketing-solutions/campaigns/search` - records path `data`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100;
  computed output fields `advertiserId`, `goal`, `name`, `objective`, `spendLimit`.
- `audiences`: GET `/2024-01/marketing-solutions/audiences` - records path `data`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100; computed output
  fields `advertiserId`, `description`, `name`, `nbActiveUsers`.
- `statistics`: GET `/2024-01/statistics/report` - records path `Rows`; query `currency` from
  template `{{ config.currency }}`, omitted when absent; `endDate` from template `{{ config.end_date
  }}`, omitted when absent; `startDate` from template `{{ config.start_date }}`, omitted when
  absent; offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size
  100.
- `mpo_advertisers`: GET `/2026-01/marketing-solutions/marketplace-performance-outcomes/advertisers`
  - records at response root.
- `mpo_sellers`: GET `/2026-01/marketing-solutions/marketplace-performance-outcomes/sellers` -
  records at response root.
- `mpo_budgets`: GET `/2026-01/marketing-solutions/marketplace-performance-outcomes/budgets` -
  records at response root.
- `mpo_seller_campaigns`: GET
  `/2026-01/marketing-solutions/marketplace-performance-outcomes/seller-campaigns` - records at
  response root.

## Write actions & risks

This connector is read-only. Read behavior: external Criteo Marketing Solutions API read of
advertiser, campaign, and ad spend data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 9 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=23, non_data_endpoint=1, out_of_scope=22, requires_elevated_scope=41.
