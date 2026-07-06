# Overview

Reads Google Search Console sites, sitemaps, and Search Analytics performance reports (by date,
query, page, country, and device) through the Search Console v3 REST API; submits/removes sites and
sitemaps through explicit write actions.

Readable streams: `sites`, `site_details`, `sitemaps`, `sitemap_details`,
`search_analytics_by_date`, `search_analytics_by_country`, `search_analytics_by_device`,
`search_analytics_by_page`, `search_analytics_by_query`.

Write actions: `add_site`, `delete_site`, `submit_sitemap`, `delete_sitemap`.

Service API documentation: https://developers.google.com/webmaster-tools/v1/api_reference_index.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Google OAuth 2.0 access token with Search Console read
  scope. Used only for Bearer auth; never logged. The 3-legged consent/acquisition and
  refresh-token-exchange dance is out of scope for this connector (credentials layer already owns
  it) -- see docs.md Known limits.
- `base_url` (optional, string); default `https://www.googleapis.com/webmasters/v3`; format `uri`;
  Search Console API base URL override for tests or proxies.
- `data_state` (optional, string); Search Analytics dataState filter (final or all). Omitted when
  unset, matching the API's own default.
- `end_date` (optional, string); format `date`; YYYY-MM-DD upper bound for search_analytics_*
  streams. Defaults to today (UTC) when unset.
- `feedpath` (optional, string); Sitemap feed URL/path for sitemap_details.
- `max_pages` (optional, string); Maximum pages fetched per site for search_analytics_* streams. A
  positive integer, or 'all'/'unlimited' (default) for no cap.
- `mode` (optional, string).
- `page_size` (optional, integer); default `25000`; Search Analytics rowLimit per page. Search
  Console caps this at 25000.
- `search_type` (optional, string); default `web`; Search Analytics searchType/type filter (web,
  image, video, news, discover, googleNews).
- `site_url` (optional, string); Single Search Console site property for site_details and
  sitemap_details.
- `site_urls` (optional, string); Comma- or newline-separated Search Console site properties (e.g.
  https://example.com/ or sc-domain:example.com) to fan out over for the sitemaps and
  search_analytics_* streams. Required for those streams only; sites itself is account-scoped and
  does not reference it.
- `start_date` (optional, string); default `2021-01-01`; format `date`; YYYY-MM-DD lower bound for
  search_analytics_* streams. The incremental cursor (previously-synced date), when present,
  overrides this on a repeat sync.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://www.googleapis.com/webmasters/v3`,
`page_size=25000`, `search_type=web`, `start_date=2021-01-01`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/sites`.

## Streams notes

Default pagination: single request; no pagination.

- `sites`: GET `/sites` - records path `siteEntry`; computed output fields `permission_level`,
  `site_url`.
- `site_details`: GET `/sites/{{ config.site_url }}` - single-object response; records path `.`;
  computed output fields `permission_level`, `site_url`.
- `sitemaps`: GET `/sites/{{ fanout.id }}/sitemaps` - records path `sitemap`; computed output fields
  `errors`, `is_pending`, `is_sitemaps_index`, `last_downloaded`, `last_submitted`, `path`, `type`,
  `warnings`; fan-out; ids from config field `site_urls`; id inserted into the request path; stamps
  `site_url`.
- `sitemap_details`: GET `/sites/{{ config.site_url }}/sitemaps/{{ config.feedpath }}` -
  single-object response; records path `.`; computed output fields `errors`, `is_pending`,
  `is_sitemaps_index`, `last_downloaded`, `last_submitted`, `path`, `site_url`, `type`, `warnings`.
- `search_analytics_by_date`: POST `/sites/{{ fanout.id }}/searchAnalytics/query` - records path
  `rows`; fan-out; ids from config field `site_urls`; id inserted into the request path.
- `search_analytics_by_country`: POST `/sites/{{ fanout.id }}/searchAnalytics/query` - records path
  `rows`; fan-out; ids from config field `site_urls`; id inserted into the request path.
- `search_analytics_by_device`: POST `/sites/{{ fanout.id }}/searchAnalytics/query` - records path
  `rows`; fan-out; ids from config field `site_urls`; id inserted into the request path.
- `search_analytics_by_page`: POST `/sites/{{ fanout.id }}/searchAnalytics/query` - records path
  `rows`; fan-out; ids from config field `site_urls`; id inserted into the request path.
- `search_analytics_by_query`: POST `/sites/{{ fanout.id }}/searchAnalytics/query` - records path
  `rows`; fan-out; ids from config field `site_urls`; id inserted into the request path.

## Write actions & risks

Overall write risk: adds or removes Search Console site properties and submits or deletes sitemap
resources.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `add_site`: PUT `/sites/{{ record.site_url | urlencode }}` - kind `create`; body type `none`; path
  fields `site_url`; required record fields `site_url`; accepted fields `site_url`; risk: adds a
  site property to the authenticated Search Console account.
- `delete_site`: DELETE `/sites/{{ record.site_url | urlencode }}` - kind `delete`; body type
  `none`; path fields `site_url`; required record fields `site_url`; accepted fields `site_url`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: removes a
  site property from the authenticated Search Console account.
- `submit_sitemap`: PUT `/sites/{{ record.site_url | urlencode }}/sitemaps/{{ record.feedpath |
  urlencode }}` - kind `create`; body type `none`; path fields `site_url`, `feedpath`; required
  record fields `site_url`, `feedpath`; accepted fields `feedpath`, `site_url`; risk: submits a
  sitemap URL for a Search Console site property.
- `delete_sitemap`: DELETE `/sites/{{ record.site_url | urlencode }}/sitemaps/{{ record.feedpath |
  urlencode }}` - kind `delete`; body type `none`; path fields `site_url`, `feedpath`; required
  record fields `site_url`, `feedpath`; accepted fields `feedpath`, `site_url`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: deletes a sitemap from a
  Search Console site property.

## Known limits

- Batch defaults: read_page_size=25000.
- API coverage includes 9 stream-backed endpoint group(s), 4 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=1.
