# pm connectors inspect appsflyer

```text
NAME
  pm connectors inspect appsflyer - AppsFlyer connector manual

SYNOPSIS
  pm connectors inspect appsflyer
  pm connectors inspect appsflyer --json
  pm credentials add <name> --connector appsflyer [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads AppsFlyer raw-data CSV export reports (installs, in-app events) through the AppsFlyer Pull API. Read-only.

ICON
  asset: icons/appsflyer.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://dev.appsflyer.com/hc/reference

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  app_id
  base_url
  end_date
  mode
  start_date
  timezone
  api_token (secret)

ETL STREAMS
  installs_report:
    fields: appsflyer_id(), campaign(), event_time(), media_source()
  in_app_events_report:
    fields: appsflyer_id(), campaign(), event_time(), media_source()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external AppsFlyer API read of raw installs/in-app-event export reports
  approval: none; read-only, no writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect appsflyer

  # Inspect as structured JSON
  pm connectors inspect appsflyer --json

AGENT WORKFLOW
  - Run pm connectors inspect appsflyer before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
