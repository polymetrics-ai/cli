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
  dateFrom
  dateTo
  direction
  messageType
  type
  access_token (secret) (required)

ETL STREAMS
  extensions:
    primary key: id
    fields: extension_number(string), id(string), name(string), status(string), stream(string), type(string)
  call_log:
    primary key: id
    cursor: start_time
    fields: direction(string), id(string), result(string), start_time(string), stream(string), type(string)
  messages:
    primary key: id
    cursor: creation_time
    fields: creation_time(string), direction(string), id(string), stream(string), subject(string), type(string)
  contacts:
    primary key: id
    fields: company(string), email(string), first_name(string), id(string), last_name(string), stream(string)
  devices:
    primary key: id
    fields: id(string), name(string), status(string), stream(string), type(string)

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
