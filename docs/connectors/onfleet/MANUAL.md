# pm connectors inspect onfleet

```text
NAME
  pm connectors inspect onfleet - Onfleet connector manual

SYNOPSIS
  pm connectors inspect onfleet
  pm connectors inspect onfleet --json
  pm credentials add <name> --connector onfleet [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Onfleet tasks, workers, teams, hubs, and administrators through the Onfleet REST API.

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
  max_pages
  mode
  api_key (secret)

ETL STREAMS
  tasks:
    primary key: id
    cursor: timeLastModified
    fields: completed(), creator(), executor(), id(), merchant(), shortId(), state(), timeCreated(), timeLastModified(), trackingURL(), worker()
  workers:
    primary key: id
    cursor: timeLastModified
    fields: activeTask(), id(), name(), onDuty(), phone(), timeCreated(), timeLastModified(), timeLastSeen()
  teams:
    primary key: id
    cursor: timeLastModified
    fields: hub(), id(), name(), timeCreated(), timeLastModified()
  hubs:
    primary key: id
    fields: address(), id(), name()
  administrators:
    primary key: id
    cursor: timeLastModified
    fields: email(), id(), isActive(), name(), timeCreated(), timeLastModified(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Onfleet API read of delivery task and workforce data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect onfleet

  # Inspect as structured JSON
  pm connectors inspect onfleet --json

AGENT WORKFLOW
  - Run pm connectors inspect onfleet before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
