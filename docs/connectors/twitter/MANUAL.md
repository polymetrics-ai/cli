# pm connectors inspect twitter

```text
NAME
  pm connectors inspect twitter - Twitter connector manual

SYNOPSIS
  pm connectors inspect twitter
  pm connectors inspect twitter --json
  pm credentials add <name> --connector twitter [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads tweets and their authors matching a search query from the Twitter (X) API v2 recent search endpoint using an App-only Bearer token.

ICON
  asset: icons/twitter.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.twitter.com/en/docs/twitter-api

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  end_date
  max_pages
  mode
  page_size
  query
  start_date
  api_key (secret)

ETL STREAMS
  tweets:
    primary key: id
    cursor: created_at
    fields: author_id(), conversation_id(), created_at(), id(), in_reply_to_user_id(), lang(), possibly_sensitive(), public_metrics(), source(), text()
  authors:
    primary key: id
    fields: created_at(), description(), id(), location(), name(), protected(), public_metrics(), url(), username(), verified()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Twitter (X) API read of tweets and author profiles matching a search query
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect twitter

  # Inspect as structured JSON
  pm connectors inspect twitter --json

AGENT WORKFLOW
  - Run pm connectors inspect twitter before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
