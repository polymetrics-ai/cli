# pm connectors inspect nytimes

```text
NAME
  pm connectors inspect nytimes - New York Times connector manual

SYNOPSIS
  pm connectors inspect nytimes
  pm connectors inspect nytimes --json
  pm credentials add <name> --connector nytimes [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads New York Times Most Popular (viewed, emailed, shared) articles via the NYTimes Developer APIs.

ICON
  id: nytimes
  asset: icons/nytimes.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.nytimes.com/apis

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  period
  api_key (secret)

ETL STREAMS
  most_popular_viewed:
    primary key: id
    cursor: published_date
    fields: abstract(string), byline(string), id(integer), published_date(string), section(string), source(string), title(string), type(string), updated(string), uri(string), url(string)
  most_popular_emailed:
    primary key: id
    cursor: published_date
    fields: abstract(string), byline(string), id(integer), published_date(string), section(string), source(string), title(string), type(string), updated(string), uri(string), url(string)
  most_popular_shared:
    primary key: id
    cursor: published_date
    fields: abstract(string), byline(string), id(integer), published_date(string), section(string), source(string), title(string), type(string), updated(string), uri(string), url(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external NYTimes API read of published article metadata (no PII)
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run New York Times's declared streams and reverse-ETL actions.
  Usage: pm nytimes <command> [flags]
  Read streams
  Other Commands
    api get archive v1 year month json - Documented GET /archive/v1/{year}/{month}.json (not implemented) [intent=direct_read availability=not_implemented operation=nytimes.get.archive-v1-year-month-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get books v3 lists current list json - Documented GET /books/v3/lists/current/{list}.json (not implemented) [intent=direct_read availability=not_implemented operation=nytimes.get.books-v3-lists-current-list-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get emailed period json - Documented GET /emailed/{period}.json (not implemented) [intent=direct_read availability=not_implemented operation=nytimes.get.emailed-period-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get mostpopular v2 shared period share-type json - Documented GET /mostpopular/v2/shared/{period}/{share_type}.json (not implemented) [intent=direct_read availability=not_implemented operation=nytimes.get.mostpopular-v2-shared-period-share-type-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get search v2 articlesearch-json - Documented GET /search/v2/articlesearch.json (not implemented) [intent=direct_read availability=not_implemented operation=nytimes.get.search-v2-articlesearch-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get shared period json - Documented GET /shared/{period}.json (not implemented) [intent=direct_read availability=not_implemented operation=nytimes.get.shared-period-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get shared period share-type json - Documented GET /shared/{period}/{share_type}.json (not implemented) [intent=direct_read availability=not_implemented operation=nytimes.get.shared-period-share-type-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get topstories v2 section json - Documented GET /topstories/v2/{section}.json (not implemented) [intent=direct_read availability=not_implemented operation=nytimes.get.topstories-v2-section-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get viewed period json - Documented GET /viewed/{period}.json (not implemented) [intent=direct_read availability=not_implemented operation=nytimes.get.viewed-period-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    most popular emailed list - Run the most popular emailed ETL stream [intent=etl availability=implemented stream=most_popular_emailed]; notes: discrepancy=present-in-surface-absent-from-artifact
    most popular shared list - Run the most popular shared ETL stream [intent=etl availability=implemented stream=most_popular_shared]; notes: discrepancy=present-in-surface-absent-from-artifact
    most popular viewed list - Run the most popular viewed ETL stream [intent=etl availability=implemented stream=most_popular_viewed]; notes: discrepancy=present-in-surface-absent-from-artifact

EXAMPLES
  # Inspect as a manual
  pm connectors inspect nytimes

  # Inspect as structured JSON
  pm connectors inspect nytimes --json

AGENT WORKFLOW
  - Run pm connectors inspect nytimes before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
