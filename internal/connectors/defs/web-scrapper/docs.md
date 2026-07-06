# Overview

Reads sitemap, scraping job, account, and problematic-URL metadata, and writes sitemap/scraping-job
create/update/delete mutations, through the Web Scraper Cloud API.

Readable streams: `sitemaps`, `jobs`, `sitemaps_list`, `scraping_jobs_list`, `account`,
`problematic_urls`.

Write actions: `create_sitemap`, `update_sitemap`, `delete_sitemap`, `create_scraping_job`,
`delete_scraping_job`.

Service API documentation: https://webscraper.io/documentation/web-scraper-cloud/api.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); Web Scraper Cloud API token. Sent as the api_token query
  parameter on every request; never logged.
- `base_url` (optional, string); default `https://api.webscraper.io/api/v1`; format `uri`; Web
  Scraper Cloud API base URL override for tests or proxies.
- `mode` (optional, string).
- `scraping_job_ids` (optional, string); Comma-, whitespace-, or semicolon-separated Web Scraper
  Cloud scraping job IDs to fan out over for the problematic_urls stream (one request per id).
  Required for that stream only.
- `sitemap_id_filter` (optional, string); Optional sitemap_id query parameter forwarded to the
  scraping_jobs_list stream, narrowing results to scraping jobs created from a single sitemap.
  Omitted entirely when unset (all scraping jobs are listed).

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.webscraper.io/api/v1`.

Authentication behavior:

- API key authentication in query parameter `api_token` using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/sitemap`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `sitemaps`, `jobs`, `account`; page_number: `sitemaps_list`,
`scraping_jobs_list`, `problematic_urls`.

- `sitemaps`: GET `/sitemap` - records path `data`; emits passthrough records.
- `jobs`: GET `/scraping-job` - records path `data`; emits passthrough records.
- `sitemaps_list`: GET `/sitemaps` - records path `data`; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 100; emits passthrough records.
- `scraping_jobs_list`: GET `/scraping-jobs` - records path `data`; query `sitemap_id` from template
  `{{ config.sitemap_id_filter }}`, omitted when absent; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 100; emits passthrough records.
- `account`: GET `/account` - records path `data`; emits passthrough records.
- `problematic_urls`: GET `/scraping-job/{{ fanout.id }}/problematic-urls` - records path `data`;
  page-number pagination; page parameter `page`; no page-size parameter; starts at 1; page size 100;
  fan-out; ids from config field `scraping_job_ids`; id inserted into the request path; stamps
  `scraping_job_id`; emits passthrough records.

## Write actions & risks

Overall write risk: external mutation of the caller's own sitemaps and scraping jobs;
create_scraping_job starts a real scraping run against a target site and consumes page credits from
the caller's account, and delete_sitemap/delete_scraping_job are irreversible through the API.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_sitemap`: POST `/sitemap` - kind `create`; body type `json`; required record fields `_id`,
  `startUrl`, `selectors`; accepted fields `_id`, `selectors`, `startUrl`; risk: creates a new
  sitemap (scraper configuration) in the caller's Web Scraper Cloud account; low-risk, does not
  itself start scraping any site.
- `update_sitemap`: PUT `/sitemap/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; body fields `_id`, `startUrl`, `selectors`; required record fields `id`, `_id`, `startUrl`,
  `selectors`; accepted fields `_id`, `id`, `selectors`, `startUrl`; risk: overwrites an existing
  sitemap's start URLs and selector configuration; any scraping job created from this sitemap after
  the update uses the new configuration.
- `delete_sitemap`: DELETE `/sitemap/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: permanently removes a sitemap; any scraping job history tied to it is not
  itself deleted but the configuration can no longer be reused or edited.
- `create_scraping_job`: POST `/scraping-job` - kind `create`; body type `json`; required record
  fields `sitemap_id`; accepted fields `custom_id`, `driver`, `page_load_delay`, `proxy`,
  `request_interval`, `sitemap_id`, `start_urls`; risk: starts a real scraping run against the
  sitemap's (or start_urls override's) target site(s); consumes page credits from the caller's Web
  Scraper Cloud account for every page scraped.
- `delete_scraping_job`: DELETE `/scraping-job/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; risk: permanently removes a scraping job and its scraped data; any
  already-downloaded export is unaffected, but the job's stored records become unrecoverable through
  the API.

## Known limits

- API coverage includes 6 stream-backed endpoint group(s), 5 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=3, duplicate_of=2, out_of_scope=4.
