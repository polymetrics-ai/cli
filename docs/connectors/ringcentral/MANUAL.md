# pm connectors inspect ringcentral

```text
NAME
  pm connectors inspect ringcentral - RingCentral connector manual

SYNOPSIS
  pm connectors inspect ringcentral
  pm connectors inspect ringcentral --json
  pm credentials add <name> --connector ringcentral [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads RingCentral extensions, call logs, messages, contacts, and devices through the REST API.

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
  dateFrom
  dateTo
  direction
  messageType
  type
  access_token (secret)

ETL STREAMS
  extensions:
    primary key: id
    fields: extension_number(), id(), name(), status(), stream(), type()
  call_log:
    primary key: id
    cursor: start_time
    fields: direction(), id(), result(), start_time(), stream(), type()
  messages:
    primary key: id
    cursor: creation_time
    fields: creation_time(), direction(), id(), stream(), subject(), type()
  contacts:
    primary key: id
    fields: company(), email(), first_name(), id(), last_name(), stream()
  devices:
    primary key: id
    fields: id(), name(), status(), stream(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external RingCentral API read of account extension, call-log, message, contact, and device data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect ringcentral

  # Inspect as structured JSON
  pm connectors inspect ringcentral --json

AGENT WORKFLOW
  - Run pm connectors inspect ringcentral before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
