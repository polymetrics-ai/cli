# pm connectors inspect incident-io

```text
NAME
  pm connectors inspect incident-io - Incident.io connector manual

SYNOPSIS
  pm connectors inspect incident-io
  pm connectors inspect incident-io --json
  pm credentials add <name> --connector incident-io [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads incident.io incidents, severities, incident roles, users, and follow-ups through the incident.io REST API.

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
  mode
  page_size
  api_key (secret)

ETL STREAMS
  incidents:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), mode(), name(), reference(), severity_id(), severity_name(), status_category(), status_id(), status_name(), summary(), updated_at(), visibility()
  severities:
    primary key: id
    cursor: updated_at
    fields: created_at(), description(), id(), name(), rank(), updated_at()
  incident_roles:
    primary key: id
    cursor: updated_at
    fields: created_at(), description(), id(), instructions(), name(), role_type(), shortform(), updated_at()
  users:
    primary key: id
    fields: base_role_id(), base_role_name(), email(), id(), name(), role(), slack_user_id()
  follow_ups:
    primary key: id
    cursor: updated_at
    fields: assignee_id(), assignee_name(), completed_at(), created_at(), description(), id(), incident_id(), status(), title(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external incident.io API read of incidents, severities, roles, users, and follow-ups
  approval: none; read-only source
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect incident-io

  # Inspect as structured JSON
  pm connectors inspect incident-io --json

AGENT WORKFLOW
  - Run pm connectors inspect incident-io before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
