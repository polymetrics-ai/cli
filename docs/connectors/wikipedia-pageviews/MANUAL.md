# pm connectors inspect wikipedia-pageviews

```text
NAME
  pm connectors inspect wikipedia-pageviews - Wikipedia Pageviews connector manual

SYNOPSIS
  pm connectors inspect wikipedia-pageviews
  pm connectors inspect wikipedia-pageviews --json
  pm credentials add <name> --connector wikipedia-pageviews [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Wikimedia pageview metrics for articles and top-article reports through the public Wikimedia REST API.

ICON
  id: wikipedia-pageviews
  asset: icons/wikipedia-pageviews.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://wikitech.wikimedia.org/wiki/Analytics/AQS/Pageviews

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  access
  agent
  article
  base_url
  country
  day
  end
  month
  project
  start
  year

ETL STREAMS
  pageviews:
    primary key: id
    cursor: timestamp
    fields: access(string), agent(string), article(string), granularity(string), id(string), project(string), timestamp(string), views(integer)
  top_articles:
    primary key: id
    fields: access(string), articles(array), country(string), day(string), id(string), month(string), project(string), year(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Wikimedia public API read of aggregate pageview metrics; no authentication, no PII
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Wikipedia Pageviews's declared streams and reverse-ETL actions.
  Usage: pm wikipedia-pageviews <command> [flags]
  Read streams
  Other Commands
    api get api rest-v1 metrics mediarequests per-file project referer agent file granularity start end - Documented GET /api/rest_v1/metrics/mediarequests/per-file/{project}/{referer}/{agent}/{file}/{granularity}/{start}/{end} (not implemented) [intent=direct_read availability=not_implemented operation=wikipedia-pageviews.get.api-rest-v1-metrics-mediarequests-per-file-project-referer-agent-file-granularity-start-end]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api rest-v1 metrics pageviews aggregate project access agent granularity start end - Documented GET /api/rest_v1/metrics/pageviews/aggregate/{project}/{access}/{agent}/{granularity}/{start}/{end} (not implemented) [intent=direct_read availability=not_implemented operation=wikipedia-pageviews.get.api-rest-v1-metrics-pageviews-aggregate-project-access-agent-granularity-start-end]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api rest-v1 metrics pageviews top project access year month day - Documented GET /api/rest_v1/metrics/pageviews/top/{project}/{access}/{year}/{month}/{day} (not implemented) [intent=direct_read availability=not_implemented operation=wikipedia-pageviews.get.api-rest-v1-metrics-pageviews-top-project-access-year-month-day]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get legacy pagecounts aggregate project access-site granularity start end - Documented GET /legacy/pagecounts/aggregate/{project}/{access-site}/{granularity}/{start}/{end} (not implemented) [intent=direct_read availability=not_implemented operation=wikipedia-pageviews.get.legacy-pagecounts-aggregate-project-access-site-granularity-start-end]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get pageviews aggregate project access agent granularity start end - Documented GET /pageviews/aggregate/{project}/{access}/{agent}/{granularity}/{start}/{end} (not implemented) [intent=direct_read availability=not_implemented operation=wikipedia-pageviews.get.pageviews-aggregate-project-access-agent-granularity-start-end]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get pageviews per-article project access agent article granularity start end - Documented GET /pageviews/per-article/{project}/{access}/{agent}/{article}/{granularity}/{start}/{end} (not implemented) [intent=direct_read availability=not_implemented operation=wikipedia-pageviews.get.pageviews-per-article-project-access-agent-article-granularity-start-end]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get pageviews top project access year month day - Documented GET /pageviews/top/{project}/{access}/{year}/{month}/{day} (not implemented) [intent=direct_read availability=not_implemented operation=wikipedia-pageviews.get.pageviews-top-project-access-year-month-day]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get pageviews top-by-country project access year month - Documented GET /pageviews/top-by-country/{project}/{access}/{year}/{month} (not implemented) [intent=direct_read availability=not_implemented operation=wikipedia-pageviews.get.pageviews-top-by-country-project-access-year-month]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get pageviews top-per-country country access year month day - Documented GET /pageviews/top-per-country/{country}/{access}/{year}/{month}/{day} (not implemented) [intent=direct_read availability=not_implemented operation=wikipedia-pageviews.get.pageviews-top-per-country-country-access-year-month-day]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get pageviews v3 per-editor user-central-id granularity start end - Documented GET /pageviews/v3/per_editor/{user_central_id}/{granularity}/{start}/{end} (not implemented) [intent=direct_read availability=not_implemented operation=wikipedia-pageviews.get.pageviews-v3-per-editor-user-central-id-granularity-start-end]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get pageviews v3 top-pages-per-editor user-central-id granularity start end - Documented GET /pageviews/v3/top_pages_per_editor/{user_central_id}/{granularity}/{start}/{end} (not implemented) [intent=direct_read availability=not_implemented operation=wikipedia-pageviews.get.pageviews-v3-top-pages-per-editor-user-central-id-granularity-start-end]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    pageviews list - Run the pageviews ETL stream [intent=etl availability=implemented stream=pageviews]; notes: discrepancy=present-in-surface-absent-from-artifact
    top articles list - Run the top articles ETL stream [intent=etl availability=implemented stream=top_articles]; notes: discrepancy=present-in-surface-absent-from-artifact

EXAMPLES
  # Inspect as a manual
  pm connectors inspect wikipedia-pageviews

  # Inspect as structured JSON
  pm connectors inspect wikipedia-pageviews --json

AGENT WORKFLOW
  - Run pm connectors inspect wikipedia-pageviews before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
