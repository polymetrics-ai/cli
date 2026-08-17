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
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

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
    fields: article_id(), category(), content(), country(), creator(), description(), duplicate(), image_url(), keywords(), language(), link(), pubDate(), source_id(), source_name(), source_url(), title()
  crypto:
    primary key: article_id
    cursor: pubDate
    fields: article_id(), category(), content(), country(), creator(), description(), duplicate(), image_url(), keywords(), language(), link(), pubDate(), source_id(), source_name(), source_url(), title()
  archive:
    primary key: article_id
    cursor: pubDate
    fields: article_id(), category(), content(), country(), creator(), description(), duplicate(), image_url(), keywords(), language(), link(), pubDate(), source_id(), source_name(), source_url(), title()
  sources:
    primary key: id
    fields: category(), country(), description(), icon(), id(), language(), name(), priority(), url()

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
