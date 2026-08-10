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
  api_token (secret)

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
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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

COMMAND SURFACE
  Run Web Scrapper's declared streams and reverse-ETL actions.
  Usage: pm web-scrapper <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    account list - Run the account ETL stream [intent=etl availability=implemented stream=account]
    api delete scraping-job scrapingjobid - Documented DELETE /scraping-job/{scrapingJobId} (not implemented) [intent=direct_write availability=not_implemented operation=web-scrapper.delete.scraping-job-scrapingjobid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete sitemap sitemapid - Documented DELETE /sitemap/{sitemapId} (not implemented) [intent=direct_write availability=not_implemented operation=web-scrapper.delete.sitemap-sitemapid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete sitemap sitemapid remove-tag - Documented DELETE /sitemap/{sitemapId}/remove-tag (not implemented) [intent=direct_write availability=not_implemented operation=web-scrapper.delete.sitemap-sitemapid-remove-tag]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get scraping-job id - Documented GET /scraping-job/{id} (not implemented) [intent=direct_read availability=not_implemented operation=web-scrapper.get.scraping-job-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get scraping-job id csv - Documented GET /scraping-job/{id}/csv (not implemented) [intent=direct_read availability=not_implemented operation=web-scrapper.get.scraping-job-id-csv]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get scraping-job id data-quality - Documented GET /scraping-job/{id}/data-quality (not implemented) [intent=direct_read availability=not_implemented operation=web-scrapper.get.scraping-job-id-data-quality]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get scraping-job id json - Documented GET /scraping-job/{id}/json (not implemented) [intent=direct_read availability=not_implemented operation=web-scrapper.get.scraping-job-id-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get scraping-job id xlsx - Documented GET /scraping-job/{id}/xlsx (not implemented) [intent=direct_read availability=not_implemented operation=web-scrapper.get.scraping-job-id-xlsx]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get scraping-job scrapingjobid - Documented GET /scraping-job/{scrapingJobId} (not implemented) [intent=direct_read availability=not_implemented operation=web-scrapper.get.scraping-job-scrapingjobid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get scraping-job scrapingjobid data-quality - Documented GET /scraping-job/{scrapingJobId}/data-quality (not implemented) [intent=direct_read availability=not_implemented operation=web-scrapper.get.scraping-job-scrapingjobid-data-quality]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get scraping-job scrapingjobid extension - Documented GET /scraping-job/{scrapingJobId}/{extension} (not implemented) [intent=direct_read availability=not_implemented operation=web-scrapper.get.scraping-job-scrapingjobid-extension]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get scraping-job scrapingjobid problematic-urls - Documented GET /scraping-job/{scrapingJobId}/problematic-urls (not implemented) [intent=direct_read availability=not_implemented operation=web-scrapper.get.scraping-job-scrapingjobid-problematic-urls]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get sitemap id - Documented GET /sitemap/{id} (not implemented) [intent=direct_read availability=not_implemented operation=web-scrapper.get.sitemap-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get sitemap id scheduler - Documented GET /sitemap/{id}/scheduler (not implemented) [intent=direct_read availability=not_implemented operation=web-scrapper.get.sitemap-id-scheduler]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get sitemap sitemapid - Documented GET /sitemap/{sitemapId} (not implemented) [intent=direct_read availability=not_implemented operation=web-scrapper.get.sitemap-sitemapid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get sitemap sitemapid scheduler - Documented GET /sitemap/{sitemapId}/scheduler (not implemented) [intent=direct_read availability=not_implemented operation=web-scrapper.get.sitemap-sitemapid-scheduler]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get sitemap sitemapid tags - Documented GET /sitemap/{sitemapId}/tags (not implemented) [intent=direct_read availability=not_implemented operation=web-scrapper.get.sitemap-sitemapid-tags]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post scraping-job scrapingjobid continue - Documented POST /scraping-job/{scrapingJobId}/continue (not implemented) [intent=direct_write availability=not_implemented operation=web-scrapper.post.scraping-job-scrapingjobid-continue]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post scraping-job scrapingjobid stop - Documented POST /scraping-job/{scrapingJobId}/stop (not implemented) [intent=direct_write availability=not_implemented operation=web-scrapper.post.scraping-job-scrapingjobid-stop]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post sitemap id disable-scheduler - Documented POST /sitemap/{id}/disable-scheduler (not implemented) [intent=direct_write availability=not_implemented operation=web-scrapper.post.sitemap-id-disable-scheduler]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post sitemap id enable-scheduler - Documented POST /sitemap/{id}/enable-scheduler (not implemented) [intent=direct_write availability=not_implemented operation=web-scrapper.post.sitemap-id-enable-scheduler]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post sitemap sitemapid add-tag - Documented POST /sitemap/{sitemapId}/add-tag (not implemented) [intent=direct_write availability=not_implemented operation=web-scrapper.post.sitemap-sitemapid-add-tag]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post sitemap sitemapid disable-scheduler - Documented POST /sitemap/{sitemapId}/disable-scheduler (not implemented) [intent=direct_write availability=not_implemented operation=web-scrapper.post.sitemap-sitemapid-disable-scheduler]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post sitemap sitemapid enable-scheduler - Documented POST /sitemap/{sitemapId}/enable-scheduler (not implemented) [intent=direct_write availability=not_implemented operation=web-scrapper.post.sitemap-sitemapid-enable-scheduler]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put sitemap sitemapid - Documented PUT /sitemap/{sitemapId} (not implemented) [intent=direct_write availability=not_implemented operation=web-scrapper.put.sitemap-sitemapid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put sitemap sitemapid rename - Documented PUT /sitemap/{sitemapId}/rename (not implemented) [intent=direct_write availability=not_implemented operation=web-scrapper.put.sitemap-sitemapid-rename]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    create scraping job apply - Plan and execute the create scraping job reverse-ETL action [intent=reverse_etl availability=implemented write=create_scraping_job]; approval: requires plan, preview, approval, and execute; risk: starts a real scraping run against the sitemap's (or start_urls override's) target site(s); consumes page credits from the caller's Web Scraper Cloud account for every page scraped; flags: --sitemap_id (required)
    create sitemap apply - Plan and execute the create sitemap reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_sitemap]; approval: requires plan, preview, approval, and execute; risk: creates a new sitemap (scraper configuration) in the caller's Web Scraper Cloud account; low-risk, does not itself start scraping any site; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete scraping job apply - Plan and execute the delete scraping job reverse-ETL action [intent=reverse_etl availability=implemented write=delete_scraping_job]; approval: requires plan, preview, approval, and execute; risk: permanently removes a scraping job and its scraped data; any already-downloaded export is unaffected, but the job's stored records become unrecoverable through the API; flags: --id (required)
    delete sitemap apply - Plan and execute the delete sitemap reverse-ETL action [intent=reverse_etl availability=implemented write=delete_sitemap]; approval: requires plan, preview, approval, and execute; risk: permanently removes a sitemap; any scraping job history tied to it is not itself deleted but the configuration can no longer be reused or edited; flags: --id (required)
    jobs list - Run the jobs ETL stream [intent=etl availability=implemented stream=jobs]; notes: discrepancy=present-in-surface-absent-from-artifact
    problematic urls list - Run the problematic urls ETL stream [intent=etl availability=implemented stream=problematic_urls]; notes: discrepancy=present-in-surface-absent-from-artifact
    scraping jobs list list - Run the scraping jobs list ETL stream [intent=etl availability=implemented stream=scraping_jobs_list]
    sitemaps list - Run the sitemaps ETL stream [intent=etl availability=implemented stream=sitemaps]; notes: discrepancy=present-in-surface-absent-from-artifact
    sitemaps list list - Run the sitemaps list ETL stream [intent=etl availability=implemented stream=sitemaps_list]
    update sitemap apply - Plan and execute the update sitemap reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_sitemap]; approval: requires plan, preview, approval, and execute; risk: overwrites an existing sitemap's start URLs and selector configuration; any scraping job created from this sitemap after the update uses the new configuration; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags

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
