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
  api_key (secret)

ETL STREAMS
  countries:
    primary key: id
    fields: code(), datetimeFirst(), datetimeLast(), id(), name(), parameters()
  parameters:
    primary key: id
    fields: description(), displayName(), id(), name(), units()
  locations:
    primary key: id
    fields: coordinates(), country(), datetimeFirst(), datetimeLast(), id(), isMobile(), isMonitor(), locality(), name(), owner(), provider(), sensors(), timezone()
  instruments:
    primary key: id
    fields: id(), isMonitor(), manufacturer(), name()
  manufacturers:
    primary key: id
    fields: id(), instruments(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
