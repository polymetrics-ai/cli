# pm connectors inspect openaq

```text
NAME
  pm connectors inspect openaq - OpenAQ connector manual

SYNOPSIS
  pm connectors inspect openaq
  pm connectors inspect openaq --json
  pm credentials add <name> --connector openaq [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads OpenAQ air quality reference data (countries, parameters, locations, instruments, and manufacturers) from the OpenAQ v3 REST API.

ICON
  id: openaq
  asset: icons/openaq.svg
  source: official
  review_status: official_verified
  review_url: https://docs.openaq.org/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  countries_id
  mode
  api_key (secret) (required)

ETL STREAMS
  countries:
    primary key: id
    fields: code(string), datetimeFirst(string), datetimeLast(string), id(integer), name(string), parameters(array)
  parameters:
    primary key: id
    fields: description(string), displayName(string), id(integer), name(string), units(string)
  locations:
    primary key: id
    fields: coordinates(object), country(object), datetimeFirst(object), datetimeLast(object), id(integer), isMobile(boolean), isMonitor(boolean), locality(string), name(string), owner(object), provider(object), sensors(array), timezone(string)
  instruments:
    primary key: id
    fields: id(integer), isMonitor(boolean), manufacturer(object), name(string)
  manufacturers:
    primary key: id
    fields: id(integer), instruments(array), name(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external OpenAQ API read of public air-quality reference data
  approval: none; read-only public reference API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect openaq

  # Inspect as structured JSON
  pm connectors inspect openaq --json

AGENT WORKFLOW
  - Run pm connectors inspect openaq before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
