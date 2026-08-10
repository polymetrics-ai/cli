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
  api_key (secret) (required)

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
