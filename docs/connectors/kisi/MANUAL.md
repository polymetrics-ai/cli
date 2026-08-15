# pm connectors inspect kisi

```text
NAME
  pm connectors inspect kisi - Kisi connector manual

SYNOPSIS
  pm connectors inspect kisi
  pm connectors inspect kisi --json
  pm credentials add <name> --connector kisi [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Kisi physical access-control data: members, locks, groups, users, and logins via the Kisi REST API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  mode
  page_size
  api_key (secret) (required)

ETL STREAMS
  members:
    primary key: id
    fields: access_enabled(boolean), confirmed(boolean), created_at(string), email(string), id(integer), name(string), role_id(integer), updated_at(string)
  locks:
    primary key: id
    fields: created_at(string), description(string), geofence_restriction_enabled(boolean), id(integer), name(string), online(boolean), place_id(integer), updated_at(string)
  groups:
    primary key: id
    fields: created_at(string), description(string), id(integer), login_count(integer), name(string), place_id(integer), updated_at(string)
  users:
    primary key: id
    fields: confirmed(boolean), created_at(string), email(string), id(integer), name(string), updated_at(string)
  logins:
    primary key: id
    fields: created_at(string), id(integer), last_used_at(string), name(string), type(string), updated_at(string), user_id(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Kisi API read of physical access-control data
  approval: none; read-only source
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect kisi

  # Inspect as structured JSON
  pm connectors inspect kisi --json

AGENT WORKFLOW
  - Run pm connectors inspect kisi before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
