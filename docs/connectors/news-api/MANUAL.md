# pm connectors inspect news-api

```text
NAME
  pm connectors inspect news-api - News API connector manual

SYNOPSIS
  pm connectors inspect news-api
  pm connectors inspect news-api --json
  pm credentials add <name> --connector news-api [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads articles and news sources from the News API (newsapi.org): the everything search, top headlines, and the sources directory.

ICON
  asset: icons/newsapi.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://newsapi.org/docs

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  category
  country
  domains
  end_date
  exclude_domains
  language
  search_in
  search_query
  sort_by
  sources
  start_date
  api_key (secret)

ETL STREAMS
  everything:
    primary key: url
    cursor: published_at
    fields: author(), content(), description(), published_at(), source_id(), source_name(), title(), url(), url_to_image()
  top_headlines:
    primary key: url
    cursor: published_at
    fields: author(), content(), description(), published_at(), source_id(), source_name(), title(), url(), url_to_image()
  sources:
    primary key: id
    fields: category(), country(), description(), id(), language(), name(), url()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external News API read of article and source metadata
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect news-api

  # Inspect as structured JSON
  pm connectors inspect news-api --json

AGENT WORKFLOW
  - Run pm connectors inspect news-api before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
