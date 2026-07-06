# pm connectors inspect google-calendar

```text
NAME
  pm connectors inspect google-calendar - Google Calendar connector manual

SYNOPSIS
  pm connectors inspect google-calendar
  pm connectors inspect google-calendar --json
  pm credentials add <name> --connector google-calendar [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Google Calendar calendar lists, events, settings, and access control rules through the Calendar API v3 using an OAuth2 refresh token. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

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
  calendarid
  mode
  client_id (secret)
  client_refresh_token_2 (secret)
  client_secret (secret)

ETL STREAMS
  calendar_list:
    primary key: id
    fields: accessRole(), colorId(), deleted(), description(), etag(), hidden(), id(), kind(), primary(), selected(), summary(), timeZone()
  events:
    primary key: id
    cursor: updated
    fields: attendees(), created(), creator(), description(), end(), etag(), htmlLink(), iCalUID(), id(), kind(), location(), organizer(), recurringEventId(), start(), status(), summary(), updated()
  settings:
    primary key: id
    fields: etag(), id(), kind(), value()
  acl:
    primary key: id
    fields: etag(), id(), kind(), role(), scope()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Google Calendar API reads performed by the legacy connector via a Tier-2 hook
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect google-calendar

  # Inspect as structured JSON
  pm connectors inspect google-calendar --json

AGENT WORKFLOW
  - Run pm connectors inspect google-calendar before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
