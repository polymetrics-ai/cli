# Overview

Reads Google Search Console sites, sitemap metadata, and Search Analytics performance reports (by
date, query, page, country, and device) through the official Search Console REST APIs; submits or
removes site/sitemap resources only through typed reverse-ETL write actions.

Readable streams: `sites`, `site_details`, `sitemaps`, `sitemap_details`,
`search_analytics_by_date`, `search_analytics_by_country`, `search_analytics_by_device`,
`search_analytics_by_page`, `search_analytics_by_query`.

Typed direct reads: `direct url-inspection inspect`, `direct mobile-friendly-test run`.

Write actions: `add_site`, `delete_site`, `submit_sitemap`, `delete_sitemap`.

Service API documentation: https://developers.google.com/webmaster-tools/v1/api_reference_index.
The current provider Discovery inventory and per-request-field citation research are recorded under
`.planning/phases/cli-google-search-console-parity-resume-r1/research/`.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Google OAuth 2.0 access token with Search Console
  scope. Read-only ETL and direct reads can use
  `https://www.googleapis.com/auth/webmasters.readonly`; reverse-ETL write actions require
  `https://www.googleapis.com/auth/webmasters`. Used only for Bearer auth; never logged. The
  3-legged consent/acquisition and refresh-token-exchange dance is out of scope for this connector
  (credentials layer already owns it).
- `base_url` (optional, string); default `https://searchconsole.googleapis.com`; format `uri`;
  Search Console API root URL override for tests or proxies. Supply only an origin root such as
  `https://host`, with no path, query, fragment, or user info; `http://host[:port]` is supported for
  local test proxies. Pathful values are unsupported and are not silently corrected. The connector
  appends the official `/webmasters/v3` and `/v1` prefixes.
- `site_urls` (optional, string); comma- or newline-separated Search Console site properties (for
  example `https://example.com/` or `sc-domain:example.com`) to fan out over for `sitemaps` and
  `search_analytics_*` streams. When unset, those streams fall back to `site_url`.
- `site_url` (optional, string); single Search Console site property for `site_details`,
  `sitemap_details`, the `sitemaps` and `search_analytics_*` fallback, and validation defaults.
- `feedpath` (optional, string); sitemap feed URL/path for `sitemap_details` and sitemap
  write/delete validation defaults.
- `inspection_url` and `mobile_test_url` (optional, string); fixture/live validation defaults for
  typed direct-read sweeps.
- `start_date` (optional, string); default `2021-01-01`; format `date`; lower bound for
  `search_analytics_*` streams.
- `end_date` (optional, string); format `date`; upper bound for `search_analytics_*` streams.
- `search_type` (optional, string); default `web`; stream Search Analytics type filter.
- `data_state` (optional, string); stream Search Analytics dataState filter (`final`, `all`, or
  `hourly_all`).
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
- `search_analytics_by_date`: POST `/webmasters/v3/sites/{{ fanout.id }}/searchAnalytics/query`
  with the fixed `date` dimension; fans out over `site_urls` and falls back to `site_url`.
- `search_analytics_by_country`, `search_analytics_by_device`, `search_analytics_by_page`, and
  `search_analytics_by_query`: the same POST with `date` plus the named dimension. Each record is
  therefore grouped per date and dimension value; `date` remains the incremental cursor and part
  of the primary key.

## Typed direct reads

All direct reads are fixed-target POST operations with closed request-body schemas,
`application/json`, `json_redacted` output, and a 1 MiB operation response cap. They do not accept a
raw method, path, query, or body escape hatch.

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
  validation.
- `base_url` must be an origin-only root such as `https://searchconsole.googleapis.com`.
  Credential creation and command overlays do not enforce that shape, so validate overrides before
  use; pathful values are unsupported and are not silently corrected.
- API coverage is 11/11 unique provider-published operations: five operations are reachable via
  ETL streams, four via typed reverse-ETL writes, and two additional operations via typed bounded
  direct reads. Search Analytics has five dimension-specific ETL conveniences over its single
  documented POST operation. There are 0 excluded, binary, CDC, generic raw, or planned operations.
