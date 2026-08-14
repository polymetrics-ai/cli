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
  account_id (required)
  base_url
  mode
  page_size
  personal_access_token (secret) (required)

ETL STREAMS
  contacts:
    primary key: id
    fields: custom_fields(object), email(string), first_name(string), id(integer), language(string), name(string)
  custom_fields:
    primary key: id
    fields: id(string), name(string), placeholder(string), type(string)
  interests:
    primary key: id
    fields: description(string), id(string), label(string), name(string), visibility(string)
  segments:
    primary key: id
    fields: id(string), name(string), number_of_contacts(integer)
  sources:
    primary key: id
    fields: id(integer), name(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Flexmail API read of contact and marketing-list data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Flexmail's declared streams and reverse-ETL actions.
  Usage: pm flexmail <command> [flags]
  Read streams
  Other Commands
    contacts list - Run the contacts ETL stream [intent=etl availability=implemented stream=contacts]
    custom fields list - Run the custom fields ETL stream [intent=etl availability=implemented stream=custom_fields]
    interests list - Run the interests ETL stream [intent=etl availability=implemented stream=interests]
    segments list - Run the segments ETL stream [intent=etl availability=implemented stream=segments]
    sources list - Run the sources ETL stream [intent=etl availability=implemented stream=sources]

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
