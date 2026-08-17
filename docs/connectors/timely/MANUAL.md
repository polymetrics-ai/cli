# pm connectors inspect timely

```text
NAME
  pm connectors inspect timely - Timely connector manual

SYNOPSIS
  pm connectors inspect timely
  pm connectors inspect timely --json
  pm credentials add <name> --connector timely [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads users, projects, clients, calendar/time events, time entries (hours), tags (labels), and teams from the Timely API. Read-only: every Timely mutation endpoint requires a nested single-key JSON body envelope (e.g. {"client": {...}}) the engine's declarative write dialect cannot express.

ICON
  asset: icons/timely.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://dev.timelyapp.com/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  account_id
  base_url
  start_date
  bearer_token (secret)

ETL STREAMS
  users:
    primary key: id
    fields: created_at(), email(), id(), name(), updated_at()
  projects:
    primary key: id
    fields: client_id(), created_at(), id(), name(), updated_at()
  clients:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  events:
    primary key: id
    fields: created_at(), duration(), id(), project_id(), updated_at(), user_id()
  hours:
    primary key: id
    fields: billable(), billed(), created_at(), day(), deleted(), external_id(), from(), id(), note(), project_id(), to(), uid(), updated_at(), user_id()
  labels:
    primary key: id
    fields: active(), created_at(), emoji(), external_id(), id(), name(), parent_id(), sequence(), updated_at()
  teams:
    primary key: id
    fields: color(), emoji(), external_id(), id(), name(), project_ids(), user_ids()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Timely API read of user, project, client, time event/entry, tag, and team data
  approval: none; read-only, no reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect timely

  # Inspect as structured JSON
  pm connectors inspect timely --json

AGENT WORKFLOW
  - Run pm connectors inspect timely before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
