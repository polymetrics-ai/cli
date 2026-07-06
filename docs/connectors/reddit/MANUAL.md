# pm connectors inspect reddit

```text
NAME
  pm connectors inspect reddit - Reddit connector manual

SYNOPSIS
  pm connectors inspect reddit
  pm connectors inspect reddit --json
  pm credentials add <name> --connector reddit [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads subreddit posts and comments through the Reddit OAuth API listing endpoints.

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
  subreddit
  access_token (secret)

ETL STREAMS
  posts:
    primary key: id
    cursor: created_utc
    fields: author(), created_utc(), id(), name(), permalink(), subreddit(), title()
  comments:
    primary key: id
    cursor: created_utc
    fields: author(), body(), created_utc(), id(), name(), permalink(), subreddit()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Reddit OAuth API read of public subreddit posts and comments
  approval: none; read-only, caller-supplied OAuth token
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect reddit

  # Inspect as structured JSON
  pm connectors inspect reddit --json

AGENT WORKFLOW
  - Run pm connectors inspect reddit before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
