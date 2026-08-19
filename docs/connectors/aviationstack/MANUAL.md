# pm connectors inspect aviationstack

```text
NAME
  pm connectors inspect aviationstack - Aviationstack connector manual

SYNOPSIS
  pm connectors inspect aviationstack
  pm connectors inspect aviationstack --json
  pm credentials add <name> --connector aviationstack [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads aviationstack flights and aviation reference data (airlines, airports, airplanes, countries) through the aviationstack REST API. Read-only.

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
  mode
  access_key (secret) (required)

ETL STREAMS
  flights:
    primary key: flight_date, flight_iata
    cursor: flight_date
    fields: airline_iata(string), airline_name(string), arrival_airport(string), arrival_iata(string), arrival_scheduled(string), departure_airport(string), departure_iata(string), departure_scheduled(string), flight_date(string), flight_iata(string), flight_icao(string), flight_number(string), flight_status(string)
  airlines:
    primary key: id
    fields: airline_name(string), callsign(string), country_iso2(string), country_name(string), date_founded(string), fleet_size(string), iata_code(string), icao_code(string), id(string), status(string), type(string)
  airports:
    primary key: id
    fields: airport_name(string), city_iata_code(string), country_iso2(string), country_name(string), gmt(string), iata_code(string), icao_code(string), id(string), latitude(string), longitude(string), timezone(string)
  airplanes:
    primary key: id
    fields: airline_iata_code(string), first_flight_date(string), iata_type(string), icao_code_hex(string), id(string), model_code(string), model_name(string), plane_owner(string), plane_status(string), production_line(string), registration_number(string)
  countries:
    primary key: id
    fields: capital(string), continent(string), country_iso2(string), country_iso3(string), country_iso_numeric(string), country_name(string), currency_code(string), id(string), phone_prefix(string), population(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external aviationstack API read of flight and aviation reference data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect aviationstack

  # Inspect as structured JSON
  pm connectors inspect aviationstack --json

AGENT WORKFLOW
  - Run pm connectors inspect aviationstack before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
