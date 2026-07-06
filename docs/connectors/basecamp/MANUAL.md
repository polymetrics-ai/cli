# pm connectors inspect basecamp

```text
NAME
  pm connectors inspect basecamp - Basecamp connector manual

SYNOPSIS
  pm connectors inspect basecamp
  pm connectors inspect basecamp --json
  pm credentials add <name> --connector basecamp [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Basecamp 3 projects, people, and account activity events through the Basecamp REST API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

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
  account_id
  base_url
  mode
  start_date
  client_id (secret)
  client_refresh_token_2 (secret)
  client_secret (secret)

ETL STREAMS
  projects:
    primary key: id
    cursor: updated_at
    fields: app_url(), bookmark_url(), created_at(), description(), id(), name(), purpose(), status(), updated_at(), url()
  people:
    primary key: id
    cursor: updated_at
    fields: admin(), client(), created_at(), email_address(), id(), name(), owner(), personable_type(), time_zone(), title(), updated_at()
  events:
    primary key: id
    cursor: created_at
    fields: action(), created_at(), id(), kind(), recording_id(), summary()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Basecamp API reads performed by the legacy connector via a Tier-2 hook
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect basecamp

  # Inspect as structured JSON
  pm connectors inspect basecamp --json

AGENT WORKFLOW
  - Run pm connectors inspect basecamp before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
