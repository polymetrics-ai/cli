# pm connectors inspect weatherstack

```text
NAME
  pm connectors inspect weatherstack - Weatherstack connector manual

SYNOPSIS
  pm connectors inspect weatherstack
  pm connectors inspect weatherstack --json
  pm credentials add <name> --connector weatherstack [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads current, historical, forecast, marine, and location-autocomplete weather data from Weatherstack. Read-only.

ICON
  asset: icons/weatherstack.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://weatherstack.com/documentation

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  autocomplete_query
  base_url
  forecast_days
  historical_date
  language
  latitude
  longitude
  mode
  query
  units
  access_key (secret)

ETL STREAMS
  current:
    primary key: id
    fields: current(), id(), location()
  historical:
    primary key: id
    fields: historical(), id(), location()
  forecast:
    primary key: id
    fields: forecast(), id(), location()
  marine:
    primary key: id
    fields: current(), id(), location()
  autocomplete:
    primary key: name, region, country, lat, lon
    fields: country(), lat(), lon(), name(), region(), timezone_id(), utc_offset()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Weatherstack API read of public weather data
  approval: none; read-only public weather data connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect weatherstack

  # Inspect as structured JSON
  pm connectors inspect weatherstack --json

AGENT WORKFLOW
  - Run pm connectors inspect weatherstack before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
