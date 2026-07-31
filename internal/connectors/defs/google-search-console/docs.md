# Overview

Reads Google Search Console sites, sitemap metadata, and Search Analytics performance reports (by
date, query, page, country, and device) through the official Search Console REST APIs; submits or
removes site/sitemap resources only through typed reverse-ETL write actions.

Readable streams: `sites`, `site_details`, `sitemaps`, `sitemap_details`,
`search_analytics_by_date`, `search_analytics_by_country`, `search_analytics_by_device`,
`search_analytics_by_page`, `search_analytics_by_query`.

Typed direct reads: `direct search-analytics query`, `direct url-inspection inspect`,
`direct mobile-friendly-test run`.

Write actions: `add_site`, `delete_site`, `submit_sitemap`, `delete_sitemap`.

Service API documentation: https://developers.google.com/webmaster-tools/v1/api_reference_index.
Official source audit for this checkpoint is archived under
`.planning/phases/issue-3038-google-search-console-parity-wave03/research/`.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Google OAuth 2.0 access token with Search Console
  scope. Read-only ETL and direct reads can use
  `https://www.googleapis.com/auth/webmasters.readonly`; reverse-ETL write actions require
  `https://www.googleapis.com/auth/webmasters`. Used only for Bearer auth; never logged. The
  3-legged consent/acquisition and refresh-token-exchange dance is out of scope for this connector
  (credentials layer already owns it).
- `base_url` (optional, string); default `https://searchconsole.googleapis.com`; format `uri`;
  Search Console API base URL override for tests or proxies. The connector appends the official
  `/webmasters/v3` and `/v1` endpoint prefixes and avoids duplicating `/webmasters/v3` when a
  legacy base URL already includes it.
- `site_urls` (optional, string); comma- or newline-separated Search Console site properties (for
  example `https://example.com/` or `sc-domain:example.com`) to fan out over for `sitemaps` and
  `search_analytics_*` streams.
- `site_url` (optional, string); single Search Console site property for `site_details`,
  `sitemap_details`, and certification defaults.
- `feedpath` (optional, string); sitemap feed URL/path for `sitemap_details`.
- `inspection_url` and `mobile_test_url` (optional, string); fixture/live certification defaults for
  typed direct-read sweeps.
- `start_date` (optional, string); default `2021-01-01`; format `date`; lower bound for
  `search_analytics_*` streams.
- `end_date` (optional, string); format `date`; upper bound for `search_analytics_*` streams.
- `search_type` (optional, string); default `web`; stream Search Analytics type filter.
- `data_state` (optional, string); stream Search Analytics dataState filter.
- `page_size` (optional, integer); default `25000`; Search Analytics rowLimit per stream page.
- `max_pages` (optional, string); positive integer or `all`/`unlimited` for stream pagination.
- `mode` (optional, string); `live` or `fixture`.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://searchconsole.googleapis.com`, `page_size=25000`,
`search_type=web`, `start_date=2021-01-01`.

Connection checks call GET `/webmasters/v3/sites`.

## Streams notes

Default pagination: single request; Search Analytics streams page through a connector-owned hook
because the official operation is a POST whose cursor (`startRow`) lives inside the JSON body.

- `sites`: GET `/webmasters/v3/sites` - records path `siteEntry`.
- `site_details`: GET `/webmasters/v3/sites/{{ config.site_url }}` - single-object response.
- `sitemaps`: GET `/webmasters/v3/sites/{{ fanout.id }}/sitemaps` - records path `sitemap`; fans
  out over `site_urls` and falls back to `site_url` in the hook.
- `sitemap_details`: GET `/webmasters/v3/sites/{{ config.site_url }}/sitemaps/{{ config.feedpath }}`.
- `search_analytics_by_*`: POST `/webmasters/v3/sites/{{ fanout.id }}/searchAnalytics/query` with
  fixed stream dimensions (`date`, `country`, `device`, `page`, or `query`).

## Typed direct reads

All direct reads are fixed-target POST operations with closed request-body schemas,
`application/json`, `json_redacted` output, and a 1 MiB operation response cap. They do not accept a
raw method, path, query, or body escape hatch.

- `direct search-analytics query`: POST
  `/webmasters/v3/sites/{siteUrl}/searchAnalytics/query`; supports required `--site-url`,
  `--start-date`, and `--end-date`, optional bounded dimension slots `--dimension-1` through
  `--dimension-5`, `--type`, `--data-state`, and `--aggregation-type`. The first page is fixed to
  `rowLimit=25000`, `startRow=0`.
- `direct url-inspection inspect`: POST `/v1/urlInspection/index:inspect`; supports
  `--inspection-url`, `--site-url`, and optional `--language-code`.
- `direct mobile-friendly-test run`: POST `/v1/urlTestingTools/mobileFriendlyTest:run`; supports
  `--url` and optional `--request-screenshot`.

## Write actions & risks

Overall write risk: adds or removes Search Console site properties and submits or deletes sitemap
resources. Reverse ETL writes must follow plan -> preview -> approval -> execute. Destructive
operations also require typed confirmation. No write action accepts a raw body.

- `add_site`: PUT `/webmasters/v3/sites/{{ record.site_url | urlencode }}` - kind `create`; record
  schema is closed over required `site_url`; redacts `site_url`.
- `delete_site`: DELETE `/webmasters/v3/sites/{{ record.site_url | urlencode }}` - kind `delete`;
  record schema is closed over required `site_url`; redacts `site_url`; idempotent when the provider
  returns HTTP 404; confirmation `destructive`.
- `submit_sitemap`: PUT
  `/webmasters/v3/sites/{{ record.site_url | urlencode }}/sitemaps/{{ record.feedpath | urlencode }}`
  - kind `create`; record schema is closed over required `site_url` and `feedpath`; redacts both.
- `delete_sitemap`: DELETE
  `/webmasters/v3/sites/{{ record.site_url | urlencode }}/sitemaps/{{ record.feedpath | urlencode }}`
  - kind `delete`; record schema is closed over required `site_url` and `feedpath`; redacts both;
  idempotent when the provider returns HTTP 404; confirmation `destructive`.

## Known limits

- Fixture and local validation only in this wave; this file does not claim live provider
  certification.
- API coverage is 11/11 documented operations from the official audit: 4 ETL read endpoint groups,
  4 reverse-ETL writes, and 3 typed bounded direct reads. There are 0 excluded, binary, CDC, or
  generic raw operations.
