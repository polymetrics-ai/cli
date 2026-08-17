# pm connectors inspect facebook-pages

```text
NAME
  pm connectors inspect facebook-pages - Facebook Pages connector manual

SYNOPSIS
  pm connectors inspect facebook-pages
  pm connectors inspect facebook-pages --json
  pm credentials add <name> --connector facebook-pages [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Facebook Page metadata and posts from the Graph API. Read-only.

ICON
  asset: icons/facebook.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.facebook.com/docs/pages/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  page_id
  page_size
  access_token (secret)

ETL STREAMS
  page:
    primary key: id
    fields: category(), fan_count(), id(), link(), name()
  posts:
    primary key: id
    cursor: updated_time
    fields: created_time(), id(), message(), permalink_url(), updated_time()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Facebook Graph API read of page metadata and posts
  approval: none; read-only, no writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect facebook-pages

  # Inspect as structured JSON
  pm connectors inspect facebook-pages --json

AGENT WORKFLOW
  - Run pm connectors inspect facebook-pages before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
