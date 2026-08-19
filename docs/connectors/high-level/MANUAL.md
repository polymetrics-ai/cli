# pm connectors inspect high-level

```text
NAME
  pm connectors inspect high-level - High Level connector manual

SYNOPSIS
  pm connectors inspect high-level
  pm connectors inspect high-level --json
  pm credentials add <name> --connector high-level [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads HighLevel (Go HighLevel / LeadConnector) contacts, opportunities, pipelines, custom fields, and form submissions for a location through the HighLevel REST API.

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
  api_version
  base_url
  location_id (required)
  api_key (secret) (required)

ETL STREAMS
  pipelines:
    primary key: id
    fields: dateAdded(string), dateUpdated(string), id(string), locationId(string), name(string), stages(array)
  contacts:
    primary key: id
    cursor: dateUpdated
    fields: contactName(string), dateAdded(string), dateUpdated(string), email(string), firstName(string), id(string), lastName(string), locationId(string), phone(string), source(string), type(string)
  opportunities:
    primary key: id
    cursor: dateUpdated
    fields: assignedTo(string), contactId(string), dateAdded(string), dateUpdated(string), id(string), monetaryValue(number), name(string), pipelineId(string), pipelineStageId(string), source(string), status(string)
  custom_fields:
    primary key: id
    fields: dataType(string), fieldKey(string), id(string), model(string), name(string), position(integer)
  form_submissions:
    primary key: id
    cursor: createdAt
    fields: contactId(string), createdAt(string), email(string), formId(string), id(string), locationId(string), name(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external HighLevel (LeadConnector) API read of contact, opportunity, pipeline, custom field, and form submission data for a configured location
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect high-level

  # Inspect as structured JSON
  pm connectors inspect high-level --json

AGENT WORKFLOW
  - Run pm connectors inspect high-level before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
