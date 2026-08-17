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
  api_token (secret)

ETL STREAMS
  alerts:
    primary key: id
    cursor: created_at
    fields: alias(), created_at(), details(), id(), last_occurred_at(), message(), owner(), priority(), responders(), source(), status(), tags(), tiny_id(), updated_at()
  incidents:
    primary key: id
    cursor: created_at
    fields: created_at(), description(), id(), impacted_services(), message(), owner_team(), priority(), responders(), status(), tags(), tiny_id(), updated_at()
  users:
    primary key: id
    fields: blocked(), full_name(), id(), locale(), role(), time_zone(), username(), verified()
  teams:
    primary key: id
    fields: created_at(), description(), id(), members(), name(), updated_at()
  services:
    primary key: id
    fields: created_at(), description(), id(), name(), tags(), team_id(), updated_at(), visibility()

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
