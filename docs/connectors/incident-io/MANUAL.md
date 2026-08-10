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
  mode
  page_size
  api_key (secret) (required)

ETL STREAMS
  incidents:
    primary key: id
    cursor: updated_at
    fields: created_at(string), id(string), mode(string), name(string), reference(string), severity_id(string), severity_name(string), status_category(string), status_id(string), status_name(string), summary(string), updated_at(string), visibility(string)
  severities:
    primary key: id
    cursor: updated_at
    fields: created_at(string), description(string), id(string), name(string), rank(integer), updated_at(string)
  incident_roles:
    primary key: id
    cursor: updated_at
    fields: created_at(string), description(string), id(string), instructions(string), name(string), role_type(string), shortform(string), updated_at(string)
  users:
    primary key: id
    fields: base_role_id(string), base_role_name(string), email(string), id(string), name(string), role(string), slack_user_id(string)
  follow_ups:
    primary key: id
    cursor: updated_at
    fields: assignee_id(string), assignee_name(string), completed_at(string), created_at(string), description(string), id(string), incident_id(string), status(string), title(string), updated_at(string)

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
