# pm connectors inspect factorial

```text
NAME
  pm connectors inspect factorial - Factorial connector manual

SYNOPSIS
  pm connectors inspect factorial
  pm connectors inspect factorial --json
  pm credentials add <name> --connector factorial [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads FactorialHR employees, teams, time-off leaves, leave types, and locations through the Factorial REST API.

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
  mode
  api_key (secret) (required)

ETL STREAMS
  employees:
    primary key: id
    cursor: updated_at
    fields: active(boolean), birthday_on(string), company_id(integer), created_at(string), email(string), first_name(string), full_name(string), gender(string), id(integer), last_name(string), legal_entity_id(integer), location_id(integer), manager_id(integer), team_ids(array), terminated_on(string), updated_at(string)
  teams:
    primary key: id
    fields: avatar(string), company_id(integer), description(string), employee_ids(array), id(integer), lead_ids(array), name(string)
  leaves:
    primary key: id
    cursor: updated_at
    fields: approved(boolean), created_at(string), description(string), employee_id(integer), finish_on(string), half_day(string), id(integer), leave_type_id(integer), start_on(string), updated_at(string)
  leave_types:
    primary key: id
    fields: active(boolean), approval_required(boolean), color(string), company_id(integer), id(integer), identifier(string), name(string)
  locations:
    primary key: id
    fields: address_line_1(string), city(string), company_id(integer), country(string), id(integer), main(boolean), name(string), postal_code(string), state(string), timezone(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Factorial API read of employee, team, and time-off data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Factorial's declared typed write actions.
  Usage: pm factorial <command> [flags]

SYNC TRANSPORT
  Source transport: declared
  Destination transport: unsupported
  A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.
  Source executor: declarative_api/declarative_stream_source

EXAMPLES
  # Inspect as a manual
  pm connectors inspect factorial

  # Inspect as structured JSON
  pm connectors inspect factorial --json

AGENT WORKFLOW
  - Run pm connectors inspect factorial before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
