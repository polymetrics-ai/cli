# pm connectors inspect fulcrum

```text
NAME
  pm connectors inspect fulcrum - Fulcrum connector manual

SYNOPSIS
  pm connectors inspect fulcrum
  pm connectors inspect fulcrum --json
  pm credentials add <name> --connector fulcrum [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Fulcrum forms, records, projects, choice lists, and classification sets through the Fulcrum REST API v2.

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
  page_size
  api_key (secret) (required)

ETL STREAMS
  forms:
    primary key: id
    cursor: updated_at
    fields: auto_assign(boolean), created_at(string), description(string), id(string), name(string), record_count(integer), status(string), updated_at(string)
  records:
    primary key: id
    cursor: updated_at
    fields: created_at(string), created_by(string), form_id(string), id(string), latitude(number), longitude(number), project_id(string), status(string), updated_at(string), updated_by(string)
  projects:
    primary key: id
    cursor: updated_at
    fields: created_at(string), description(string), id(string), name(string), updated_at(string)
  choice_lists:
    primary key: id
    cursor: updated_at
    fields: created_at(string), description(string), id(string), name(string), updated_at(string)
  classification_sets:
    primary key: id
    cursor: updated_at
    fields: created_at(string), description(string), id(string), name(string), updated_at(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Fulcrum API read of form, record, and project data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect fulcrum

  # Inspect as structured JSON
  pm connectors inspect fulcrum --json

AGENT WORKFLOW
  - Run pm connectors inspect fulcrum before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
