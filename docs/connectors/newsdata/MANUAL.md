# pm connectors inspect newsdata

```text
NAME
  pm connectors inspect newsdata - Newsdata connector manual

SYNOPSIS
  pm connectors inspect newsdata
  pm connectors inspect newsdata --json
  pm credentials add <name> --connector newsdata [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads latest news, cryptocurrency news, and news sources from the NewsData.io REST API.

ICON
  asset: icons/source-newsdata.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://newsdata.io/documentation

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  category
  country
  domain
  language
  query
  query_in_title
  size
  api_key (secret)

ETL STREAMS
  latest:
    primary key: article_id
    cursor: pubDate
    fields: article_id(), category(), content(), country(), creator(), description(), image_url(), keywords(), language(), link(), pubDate(), source_id(), source_priority(), title()
  crypto:
    primary key: article_id
    cursor: pubDate
    fields: article_id(), category(), content(), country(), creator(), description(), image_url(), keywords(), language(), link(), pubDate(), source_id(), source_priority(), title()
  sources:
    primary key: id
    fields: category(), country(), description(), icon(), id(), language(), name(), url()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external NewsData.io API read of article and source metadata
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect newsdata

  # Inspect as structured JSON
  pm connectors inspect newsdata --json

AGENT WORKFLOW
  - Run pm connectors inspect newsdata before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
