# pm connectors inspect opsgenie

```text
NAME
  pm connectors inspect opsgenie - Opsgenie connector manual

SYNOPSIS
  pm connectors inspect opsgenie
  pm connectors inspect opsgenie --json
  pm credentials add <name> --connector opsgenie [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Opsgenie alerts, incidents, users, teams, and services through the Opsgenie REST API.

ICON
  id: source-opsgenie
  asset: icons/source-opsgenie.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.opsgenie.com/docs/api-overview

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_token (secret) (required)

ETL STREAMS
  alerts:
    primary key: id
    cursor: created_at
    fields: alias(string), created_at(string), details(object), id(string), last_occurred_at(string), message(string), owner(string), priority(string), responders(array), source(string), status(string), tags(array), tiny_id(string), updated_at(string)
  incidents:
    primary key: id
    cursor: created_at
    fields: created_at(string), description(string), id(string), impacted_services(array), message(string), owner_team(object), priority(string), responders(array), status(string), tags(array), tiny_id(string), updated_at(string)
  users:
    primary key: id
    fields: blocked(boolean), full_name(string), id(string), locale(string), role(object), time_zone(string), username(string), verified(boolean)
  teams:
    primary key: id
    fields: created_at(string), description(string), id(string), members(array), name(string), updated_at(string)
  services:
    primary key: id
    fields: created_at(string), description(string), id(string), name(string), tags(array), team_id(string), updated_at(string), visibility(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Opsgenie API read of alerting/incident/team data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect opsgenie

  # Inspect as structured JSON
  pm connectors inspect opsgenie --json

AGENT WORKFLOW
  - Run pm connectors inspect opsgenie before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
