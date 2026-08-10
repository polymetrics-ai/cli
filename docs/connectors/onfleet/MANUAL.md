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
  api_key (secret) (required)

ETL STREAMS
  tasks:
    primary key: id
    cursor: timeLastModified
    fields: completed(boolean), creator(string), executor(string), id(string), merchant(string), shortId(string), state(integer), timeCreated(integer), timeLastModified(integer), trackingURL(string), worker(string)
  workers:
    primary key: id
    cursor: timeLastModified
    fields: activeTask(string), id(string), name(string), onDuty(boolean), phone(string), timeCreated(integer), timeLastModified(integer), timeLastSeen(integer)
  teams:
    primary key: id
    cursor: timeLastModified
    fields: hub(string), id(string), name(string), timeCreated(integer), timeLastModified(integer)
  hubs:
    primary key: id
    fields: address(string), id(string), name(string)
  administrators:
    primary key: id
    cursor: timeLastModified
    fields: email(string), id(string), isActive(boolean), name(string), timeCreated(integer), timeLastModified(integer), type(string)

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
