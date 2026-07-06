# pm connectors inspect gnews

```text
NAME
  pm connectors inspect gnews - GNews connector manual

SYNOPSIS
  pm connectors inspect gnews
  pm connectors inspect gnews --json
  pm credentials add <name> --connector gnews [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads GNews articles from the keyword search and top-headlines endpoints of the GNews REST API. Read-only.

ICON
  asset: icons/gnews.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://gnews.io/docs/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  country
  end_date
  in
  language
  max_pages
  mode
  nullable
  page_size
  query
  sortby
  start_date
  top_headlines_query
  top_headlines_topic
  api_key (secret)

ETL STREAMS
  search:
    primary key: id
    cursor: published_at
    fields: content(), description(), id(), image(), lang(), published_at(), source_country(), source_id(), source_name(), source_url(), title(), url()
  top_headlines:
    primary key: id
    cursor: published_at
    fields: content(), description(), id(), image(), lang(), published_at(), source_country(), source_id(), source_name(), source_url(), title(), url()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external GNews API read of news article search results
  approval: none; read-only news search API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect gnews

  # Inspect as structured JSON
  pm connectors inspect gnews --json

AGENT WORKFLOW
  - Run pm connectors inspect gnews before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
