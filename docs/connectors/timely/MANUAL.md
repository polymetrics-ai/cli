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
  id: timely
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
  account_id (required)
  base_url
  start_date
  bearer_token (secret) (required)

ETL STREAMS
  users:
    primary key: id
    fields: created_at(string), email(string), id(string), name(string), updated_at(string)
  projects:
    primary key: id
    fields: client_id(string), created_at(string), id(string), name(string), updated_at(string)
  clients:
    primary key: id
    fields: created_at(string), id(string), name(string), updated_at(string)
  events:
    primary key: id
    fields: created_at(string), duration(string), id(string), project_id(string), updated_at(string), user_id(string)
  hours:
    primary key: id
    fields: billable(boolean), billed(boolean), created_at(integer), day(string), deleted(boolean), external_id(string), from(string), id(integer), note(string), project_id(integer), to(string), uid(string), updated_at(integer), user_id(integer)
  labels:
    primary key: id
    fields: active(boolean), created_at(string), emoji(string), external_id(string), id(integer), name(string), parent_id(integer), sequence(integer), updated_at(string)
  teams:
    primary key: id
    fields: color(string), emoji(string), external_id(string), id(integer), name(string), project_ids(array), user_ids(array)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

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
