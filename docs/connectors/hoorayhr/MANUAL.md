# pm connectors inspect hoorayhr

```text
NAME
  pm connectors inspect hoorayhr - HoorayHR connector manual

SYNOPSIS
  pm connectors inspect hoorayhr
  pm connectors inspect hoorayhr --json
  pm credentials add <name> --connector hoorayhr [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads HoorayHR users, time-off, leave-types, and sick-leave records through the HoorayHR REST API using session-token authentication.

ICON
  id: hoorayhr
  asset: icons/hoorayhr.svg
  source: official
  review_status: official_verified
  review_url: https://api.hoorayhr.io/swagger.json

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  hoorayhrusername (required)
  mode
  hoorayhrpassword (secret) (required)

ETL STREAMS
  users:
    primary key: id
    fields: companyId(integer), companyStartDate(string), createdAt(string), email(string), firstName(string), id(integer), isAdmin(boolean), jobTitle(string), lastName(string), status(string), updatedAt(string)
  time_off:
    primary key: id
    fields: createdAt(string), end(string), id(integer), leaveTypeId(integer), leaveUnit(string), notes(string), start(string), status(string), timeOffType(string), updatedAt(string), userId(integer)
  leave_types:
    primary key: id
    fields: budget(number), color(string), createdAt(string), default(boolean), icon(string), id(integer), leaveInDays(boolean), name(string), unpaidLeave(boolean), updatedAt(string)
  sick_leaves:
    primary key: id
    fields: actualReturn(string), actualStart(string), createdAt(string), id(integer), notes(string), percentage(number), reportedReturn(string), reportedStart(string), status(string), updatedAt(string), userId(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external HoorayHR API read of employee, time-off, leave-type, and sick-leave data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect hoorayhr

  # Inspect as structured JSON
  pm connectors inspect hoorayhr --json

AGENT WORKFLOW
  - Run pm connectors inspect hoorayhr before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
