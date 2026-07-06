# pm connectors inspect rss

```text
NAME
  pm connectors inspect rss - RSS connector manual

SYNOPSIS
  pm connectors inspect rss
  pm connectors inspect rss --json
  pm credentials add <name> --connector rss [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads RSS channel metadata and feed items from any RSS 2.0 feed URL. Read-only and credential-free.

ICON
  asset: icons/rss.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  feed_url

ETL STREAMS
  items:
    primary key: id
    cursor: published_at
    fields: description(), id(), link(), published_at(), title()
  channel:
    primary key: id
    fields: description(), id(), link(), title(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external RSS feed read (XML over HTTP/HTTPS)
  approval: none; read-only, credential-free feed reader
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect rss

  # Inspect as structured JSON
  pm connectors inspect rss --json

AGENT WORKFLOW
  - Run pm connectors inspect rss before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
