# pm connectors inspect web-scrapper

```text
NAME
  pm connectors inspect web-scrapper - Web Scrapper connector manual

SYNOPSIS
  pm connectors inspect web-scrapper
  pm connectors inspect web-scrapper --json
  pm credentials add <name> --connector web-scrapper [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads sitemap, scraping job, account, and problematic-URL metadata, and writes sitemap/scraping-job create/update/delete mutations, through the Web Scraper Cloud API.

ICON
  id: web-scraper
  asset: icons/web-scraper.svg
  source: official
  review_status: official_verified
  review_url: https://webscraper.io/documentation/api

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  scraping_job_ids
  sitemap_id_filter
  api_token (secret) (required)

ETL STREAMS
  sitemaps:
    primary key: id
    fields: id(integer), name(string), url(string)
  jobs:
    primary key: id
    fields: id(integer), sitemap_id(integer), status(string)
  sitemaps_list:
    primary key: id
    fields: id(integer), name(string)
  scraping_jobs_list:
    primary key: id
    fields: custom_id(string), driver(string), id(integer), jobs_empty(integer), jobs_executed(integer), jobs_failed(integer), jobs_scheduled(integer), page_load_delay(integer), request_interval(integer), scheduled(integer), sitemap_id(integer), sitemap_name(string), status(string), stored_record_count(integer), test_run(integer), time_created(string)
  account:
    primary key: email
    fields: email(string), firstname(string), lastname(string), page_credits(integer)
  problematic_urls:
    primary key: scraping_job_id, url
    fields: scraping_job_id(string), type(string), url(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  create_sitemap:
    endpoint: POST /sitemap
    required fields: _id, startUrl, selectors
    risk: creates a new sitemap (scraper configuration) in the caller's Web Scraper Cloud account; low-risk, does not itself start scraping any site
  update_sitemap:
    endpoint: PUT /sitemap/{{ record.id }}
    required fields: id, _id, startUrl, selectors
    risk: overwrites an existing sitemap's start URLs and selector configuration; any scraping job created from this sitemap after the update uses the new configuration
  delete_sitemap:
    endpoint: DELETE /sitemap/{{ record.id }}
    required fields: id
    risk: permanently removes a sitemap; any scraping job history tied to it is not itself deleted but the configuration can no longer be reused or edited
  create_scraping_job:
    endpoint: POST /scraping-job
    required fields: sitemap_id
    risk: starts a real scraping run against the sitemap's (or start_urls override's) target site(s); consumes page credits from the caller's Web Scraper Cloud account for every page scraped
  delete_scraping_job:
    endpoint: DELETE /scraping-job/{{ record.id }}
    required fields: id
    risk: permanently removes a scraping job and its scraped data; any already-downloaded export is unaffected, but the job's stored records become unrecoverable through the API

SECURITY
  read risk: external Web Scraper Cloud API read of the caller's own sitemap/scraping-job/account/problematic-URL metadata
  write risk: external mutation of the caller's own sitemaps and scraping jobs; create_scraping_job starts a real scraping run against a target site and consumes page credits from the caller's account, and delete_sitemap/delete_scraping_job are irreversible through the API
  approval: required for create_scraping_job (consumes billable page credits and issues real outbound requests to a third-party site) and for delete_sitemap/delete_scraping_job (irreversible); create_sitemap/update_sitemap are low-risk (configuration only, no outbound scraping)
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect web-scrapper

  # Inspect as structured JSON
  pm connectors inspect web-scrapper --json

AGENT WORKFLOW
  - Run pm connectors inspect web-scrapper before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
