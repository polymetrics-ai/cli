# pm connectors inspect tvmaze-schedule

```text
NAME
  pm connectors inspect tvmaze-schedule - TVmaze Schedule connector manual

SYNOPSIS
  pm connectors inspect tvmaze-schedule
  pm connectors inspect tvmaze-schedule --json
  pm credentials add <name> --connector tvmaze-schedule [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads public TVmaze broadcast and web schedules without credentials.

ICON
  asset: icons/tvmazeschedule.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://www.tvmaze.com/api

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  base_url
  country
  date

ETL STREAMS
  schedule:
    primary key: id
    cursor: airdate
    fields: airdate(), airtime(), id(), name(), show_id(), show_name()
  web_schedule:
    primary key: id
    cursor: airdate
    fields: airdate(), airtime(), id(), name(), show_id(), show_name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external TVmaze public API read of broadcast/web schedule data
  approval: none; read-only public schedule API, no credentials
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect tvmaze-schedule

  # Inspect as structured JSON
  pm connectors inspect tvmaze-schedule --json

AGENT WORKFLOW
  - Run pm connectors inspect tvmaze-schedule before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
