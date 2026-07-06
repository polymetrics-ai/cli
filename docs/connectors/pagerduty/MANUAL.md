# pm connectors inspect pagerduty

```text
NAME
  pm connectors inspect pagerduty - PagerDuty connector manual

SYNOPSIS
  pm connectors inspect pagerduty
  pm connectors inspect pagerduty --json
  pm credentials add <name> --connector pagerduty [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads PagerDuty incidents, users, services, and teams through the REST API.

ICON
  asset: icons/pagerduty.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.pagerduty.com/api-reference/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  api_key (secret)

ETL STREAMS
  incidents:
    primary key: id
    cursor: created_at
    fields: created_at(), id(), incident_number(), status(), title()
  users:
    primary key: id
    cursor: created_at
    fields: created_at(), email(), id(), name(), role()
  services:
    primary key: id
    cursor: created_at
    fields: created_at(), description(), id(), name(), status()
  teams:
    primary key: id
    cursor: created_at
    fields: created_at(), description(), id(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external PagerDuty API read of incident, user, service, and team data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect pagerduty

  # Inspect as structured JSON
  pm connectors inspect pagerduty --json

AGENT WORKFLOW
  - Run pm connectors inspect pagerduty before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
