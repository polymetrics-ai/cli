# pm connectors inspect newsdata-io

```text
NAME
  pm connectors inspect newsdata-io - NewsData.io connector manual

SYNOPSIS
  pm connectors inspect newsdata-io
  pm connectors inspect newsdata-io --json
  pm credentials add <name> --connector newsdata-io [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads latest, crypto, and archived news articles plus available news sources from the NewsData.io REST API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  categories
  countries
  domains
  end_date
  languages
  mode
  page_size
  search_query
  start_date
  api_key (secret)

ETL STREAMS
  latest:
    primary key: article_id
    cursor: pubDate
    fields: article_id(string), category(array), content(string), country(array), creator(array), description(string), duplicate(boolean), image_url(string), keywords(array), language(string), link(string), pubDate(string), source_id(string), source_name(string), source_url(string), title(string)
  crypto:
    primary key: article_id
    cursor: pubDate
    fields: article_id(string), category(array), content(string), country(array), creator(array), description(string), duplicate(boolean), image_url(string), keywords(array), language(string), link(string), pubDate(string), source_id(string), source_name(string), source_url(string), title(string)
  archive:
    primary key: article_id
    cursor: pubDate
    fields: article_id(string), category(array), content(string), country(array), creator(array), description(string), duplicate(boolean), image_url(string), keywords(array), language(string), link(string), pubDate(string), source_id(string), source_name(string), source_url(string), title(string)
  sources:
    primary key: id
    fields: category(array), country(array), description(string), icon(string), id(string), language(array), name(string), priority(integer), url(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external NewsData.io API read of news articles and sources
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run NewsData.io's declared streams and reverse-ETL actions.
  Usage: pm newsdata-io <command> [flags]
  Read streams
  Other Commands
    api get 1 archive - Documented GET /1/archive (not implemented) [intent=direct_read availability=not_implemented operation=newsdata-io.get.1-archive]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get 1 count - Documented GET /1/count (not implemented) [intent=direct_read availability=not_implemented operation=newsdata-io.get.1-count]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get 1 crypto - Documented GET /1/crypto (not implemented) [intent=direct_read availability=not_implemented operation=newsdata-io.get.1-crypto]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get 1 crypto count - Documented GET /1/crypto/count (not implemented) [intent=direct_read availability=not_implemented operation=newsdata-io.get.1-crypto-count]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get 1 latest - Documented GET /1/latest (not implemented) [intent=direct_read availability=not_implemented operation=newsdata-io.get.1-latest]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get 1 market - Documented GET /1/market (not implemented) [intent=direct_read availability=not_implemented operation=newsdata-io.get.1-market]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get 1 market count - Documented GET /1/market/count (not implemented) [intent=direct_read availability=not_implemented operation=newsdata-io.get.1-market-count]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get 1 news - Documented GET /1/news (not implemented) [intent=direct_read availability=not_implemented operation=newsdata-io.get.1-news]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get 1 sources - Documented GET /1/sources (not implemented) [intent=direct_read availability=not_implemented operation=newsdata-io.get.1-sources]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    archive list - Run the archive ETL stream [intent=etl availability=implemented stream=archive]; notes: discrepancy=present-in-surface-absent-from-artifact
    crypto list - Run the crypto ETL stream [intent=etl availability=implemented stream=crypto]; notes: discrepancy=present-in-surface-absent-from-artifact
    latest list - Run the latest ETL stream [intent=etl availability=implemented stream=latest]; notes: discrepancy=present-in-surface-absent-from-artifact
    sources list - Run the sources ETL stream [intent=etl availability=implemented stream=sources]; notes: discrepancy=present-in-surface-absent-from-artifact

EXAMPLES
  # Inspect as a manual
  pm connectors inspect newsdata-io

  # Inspect as structured JSON
  pm connectors inspect newsdata-io --json

AGENT WORKFLOW
  - Run pm connectors inspect newsdata-io before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
