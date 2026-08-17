# pm connectors inspect flexmail

```text
NAME
  pm connectors inspect flexmail - Flexmail connector manual

SYNOPSIS
  pm connectors inspect flexmail
  pm connectors inspect flexmail --json
  pm credentials add <name> --connector flexmail [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Flexmail contacts, custom fields, interests, segments, and sources through the Flexmail REST API.

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
  account_id
  base_url
  mode
  page_size
  personal_access_token (secret)

ETL STREAMS
  contacts:
    primary key: id
    fields: custom_fields(), email(), first_name(), id(), language(), name()
  custom_fields:
    primary key: id
    fields: id(), name(), placeholder(), type()
  interests:
    primary key: id
    fields: description(), id(), label(), name(), visibility()
  segments:
    primary key: id
    fields: id(), name(), number_of_contacts()
  sources:
    primary key: id
    fields: id(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Flexmail API read of contact and marketing-list data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect flexmail

  # Inspect as structured JSON
  pm connectors inspect flexmail --json

AGENT WORKFLOW
  - Run pm connectors inspect flexmail before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
