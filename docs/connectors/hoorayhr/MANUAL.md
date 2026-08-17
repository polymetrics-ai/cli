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
  hoorayhrusername
  mode
  hoorayhrpassword (secret)

ETL STREAMS
  users:
    primary key: id
    fields: companyId(), companyStartDate(), createdAt(), email(), firstName(), id(), isAdmin(), jobTitle(), lastName(), status(), updatedAt()
  time_off:
    primary key: id
    fields: createdAt(), end(), id(), leaveTypeId(), leaveUnit(), notes(), start(), status(), timeOffType(), updatedAt(), userId()
  leave_types:
    primary key: id
    fields: budget(), color(), createdAt(), default(), icon(), id(), leaveInDays(), name(), unpaidLeave(), updatedAt()
  sick_leaves:
    primary key: id
    fields: actualReturn(), actualStart(), createdAt(), id(), notes(), percentage(), reportedReturn(), reportedStart(), status(), updatedAt(), userId()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
