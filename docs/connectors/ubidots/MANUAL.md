# pm connectors inspect ubidots

```text
NAME
  pm connectors inspect ubidots - Ubidots connector manual

SYNOPSIS
  pm connectors inspect ubidots
  pm connectors inspect ubidots --json
  pm credentials add <name> --connector ubidots [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Ubidots devices, variables, variable values, device groups, device types, dashboards, and events, and writes device/variable lifecycle mutations and new variable data points through API v2.0.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  token (secret)

ETL STREAMS
  devices:
    primary key: id
    fields: created_at(), id(), label(), name()
  variables:
    primary key: id
    fields: created_at(), id(), label(), name()
  dashboards:
    primary key: id
    fields: created_at(), id(), label(), name()
  events:
    primary key: id
    fields: created_at(), id(), label(), name()
  device_groups:
    primary key: id
    fields: created_at(), id(), label(), name()
  device_types:
    primary key: id
    fields: created_at(), id(), label(), name()
  variable_values:
    primary key: variable_id, timestamp
    fields: context(), timestamp(), value(), variable_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_device:
    endpoint: POST api/v2.0/devices/
    risk: creates a new Ubidots device; low-risk external mutation, no approval required
  update_device:
    endpoint: PATCH api/v2.0/devices/{{ record.id }}/
    required fields: id
    risk: updates the fields of an existing Ubidots device; external mutation, no approval required
  delete_device:
    endpoint: DELETE api/v2.0/devices/{{ record.id }}/
    required fields: id
    risk: permanently deletes a device and all of its variables/values; destructive and irreversible; approval required
  create_variable:
    endpoint: POST api/v2.0/variables/
    risk: creates a new variable under an existing device; low-risk external mutation, no approval required
  update_variable:
    endpoint: PATCH api/v2.0/variables/{{ record.id }}/
    required fields: id
    risk: updates the fields of an existing variable; external mutation, no approval required
  delete_variable:
    endpoint: DELETE api/v2.0/variables/{{ record.id }}/
    required fields: id
    risk: permanently deletes a variable and all of its stored values; destructive and irreversible; approval required
  create_variable_value:
    endpoint: POST api/v1.6/variables/{{ record.variable_id }}/values/
    required fields: variable_id
    optional fields: value, timestamp, context
    risk: injects a new data point (dot) into an existing variable; low-risk external mutation, no approval required

SECURITY
  read risk: external Ubidots API read of device, variable, variable-value, device-group, device-type, dashboard, and event data
  write risk: external mutation of Ubidots devices and variables (create/update/delete) and injection of new variable data points; device/variable delete is destructive and irreversible
  approval: required for delete_device/delete_variable; other writes are low-risk external mutations
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect ubidots

  # Inspect as structured JSON
  pm connectors inspect ubidots --json

AGENT WORKFLOW
  - Run pm connectors inspect ubidots before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
