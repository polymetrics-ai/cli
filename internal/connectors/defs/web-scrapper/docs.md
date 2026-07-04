# Overview

Web Scrapper is a declarative-HTTP connector for the Web Scraper Cloud API (`api.webscraper.io/api/v1`).
It reads sitemap, scraping job, account, and problematic-URL metadata, and writes sitemap/scraping-job
create/update/delete mutations. This bundle was originally migrated from
`internal/connectors/web-scrapper` (legacy wave2 fan-out: 2 read-only streams) and has since been
expanded to the full practical documented Web Scraper Cloud surface (Pass B), including its first
write actions (`capabilities.write` now `true`). The legacy package stays registered and unchanged
until wave6's registry flip.

## Auth setup

Provide a Web Scraper Cloud API token via the `api_token` secret; it is sent as the `api_token`
query parameter on every request (`auth: [{"mode": "api_key_query", "param": "api_token", ...}]`)
and is never logged. `base_url` defaults to `https://api.webscraper.io/api/v1` and may be
overridden for tests or proxies.

## Streams notes

7 streams:

- `sitemaps` (`GET /sitemap`, records at `data`) and `jobs` (`GET /scraping-job`, records at
  `data`) — unchanged from the legacy-parity migration; neither is paginated, matching legacy's
  single unparameterized request per stream.
- `sitemaps_list` (`GET /sitemaps`, records at `data`) — the real documented, paginated sitemap
  listing endpoint (`page_number` pagination, `page` query param, 100 records/page — Web Scraper
  Cloud's own documented default `per_page`). Each record is `{id, name}` (the plural endpoint's
  list shape omits the full `sitemap` selector-config JSON string that only the single-sitemap
  detail lookup returns).
- `scraping_jobs_list` (`GET /scraping-jobs`, records at `data`) — the real documented, paginated
  scraping-job listing endpoint, optionally filtered by an `sitemap_id` query parameter (wired from
  the `sitemap_id_filter` config value, omitted entirely when unset). Each record carries the full
  job-status object (`status`, `jobs_scheduled`/`jobs_executed`/`jobs_failed`/`jobs_empty`,
  `stored_record_count`, `driver`, timing fields, etc.).
- `account` (`GET /account`, records at `data`) — the caller's own account email/name/page-credit
  balance; a genuine single-object stream (primary key `email`, the only inherently-unique field
  Web Scraper Cloud's account response documents).
- `problematic_urls` (`GET /scraping-job/{id}/problematic-urls`, paginated) — fans out over the
  `scraping_job_ids` config value (comma/whitespace/semicolon-separated job IDs;
  `fan_out.ids_from.config_key`), one request per configured job id, stamping the driving id onto
  every emitted record as `scraping_job_id`. Each record is `{url, type}` (`type` is `empty` or
  `failed`, per Web Scraper Cloud's own docs); primary key is the composite
  (`scraping_job_id`, `url`) since the endpoint has no other unique per-record field.

All streams declare `"projection": "passthrough"`: legacy's `Read` emits every decoded record
verbatim (`emit(connectors.Record(item))` in `internal/connectors/web-scrapper/web_scrapper.go`,
with no field-built `connectors.Record{...}` mapping and no allowlist). Schema-mode projection
would silently drop any Web Scraper Cloud response field not enumerated in the per-stream schema
files, which is a meta-rule violation per conventions.md §8 rule 1. The schemas remain a
documentation surface listing the known/stable fields (verified against the official API
documentation's worked request/response examples); passthrough mode ensures any additional
live-response field still reaches the record instead of being silently projected away.

## Write actions & risks

`capabilities.write` is now `true`; `writes.json` declares 5 actions:

- `create_sitemap` (`POST /sitemap`) — creates a new sitemap (scraper configuration: `_id` name,
  `startUrl` array, `selectors` array). Low risk: does not itself scrape anything.
- `update_sitemap` (`PUT /sitemap/{id}`) — overwrites an existing sitemap's configuration in place.
- `delete_sitemap` (`DELETE /sitemap/{id}`) — permanently removes a sitemap; idempotent
  (`missing_ok_status: [404]`).
- `create_scraping_job` (`POST /scraping-job`) — starts a real scraping run against the sitemap's
  (or an optional `start_urls` override's) target site(s); consumes billable page credits from the
  caller's Web Scraper Cloud account for every page scraped, and issues real outbound HTTP requests
  to a third-party site chosen by the sitemap/record — flagged for approval.
- `delete_scraping_job` (`DELETE /scraping-job/{id}`) — permanently removes a scraping job and its
  stored scraped records; idempotent (`missing_ok_status: [404]`). Flagged for approval
  (irreversible data loss through the API).

See `metadata.json`'s `risk.approval` for the exact per-action approval gating.

## Known limits

- `GET /sitemap/{id}` and `GET /scraping-job/{id}` (single-resource detail lookups) are not
  separately modeled as streams: both return the identical object shape already reachable via the
  corresponding `sitemaps_list`/`scraping_jobs_list` streams' individual list records (excluded as
  `duplicate_of`), except that the single-sitemap lookup additionally includes the full `sitemap`
  selector-config JSON string — an operator needing that field syncs `sitemaps_list`'s companion
  `sitemaps`/`sitemaps_list` id alongside a targeted future single-object fan_out stream if ever
  needed; not implemented here since it would just duplicate list-stream rows otherwise.
- `GET /scraping-job/{id}/json`, `/csv`, and `/xlsx` (bulk scraped-data export/download) are
  excluded as `binary_payload`: each returns a caller-schema-defined bulk export (one selector-driven
  record shape per sitemap, not a fixed Web Scraper Cloud object shape this dialect's
  `records.path`/JSON-schema projection can describe generically).
- `GET /scraping-job/{id}/data-quality`, `GET /sitemap/{id}/scheduler`,
  `POST /sitemap/{id}/enable-scheduler`, and `POST /sitemap/{id}/disable-scheduler` are excluded as
  `out_of_scope`: none of the four appears on the official public API reference
  (`webscraper.io/documentation/web-scraper-cloud/api`) — they are exposed only via the community
  `webscraperio/api-client-php` SDK's helper methods with no published request/response field
  schema, so there is no authoritative source to build a schema from without guessing field shapes.
- **Legacy-parity discrepancy** (kept intentionally, not fixed here): the pre-existing `sitemaps`
  (`GET /sitemap`) and `jobs` (`GET /scraping-job`) streams use singular, unparameterized paths.
  The real, currently-documented Web Scraper Cloud listing endpoints are the plural `/sitemaps` and
  `/scraping-jobs` paths (added here as `sitemaps_list`/`scraping_jobs_list`) — Web Scraper Cloud
  may have renamed/pluralized these paths at some point after the legacy connector was originally
  written, or the singular paths may still work as an undocumented alias. Per the parity
  meta-rule, an existing stream's accepted-input/wire-path behavior is not changed as a Pass-B
  side effect; operators wanting the real documented paginated listing behavior should use the new
  `sitemaps_list`/`scraping_jobs_list` streams instead of `sitemaps`/`jobs`.
