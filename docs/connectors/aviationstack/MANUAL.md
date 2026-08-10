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
  access_key (secret)

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

COMMAND SURFACE
  Run Aviationstack's declared streams and reverse-ETL actions.
  Usage: pm aviationstack <command> [flags]
  Read streams
  Other Commands
    airlines list - Run the airlines ETL stream [intent=etl availability=implemented stream=airlines]
    airplanes list - Run the airplanes ETL stream [intent=etl availability=implemented stream=airplanes]
    airports list - Run the airports ETL stream [intent=etl availability=implemented stream=airports]
    api get v1 aircraft-types - Documented GET /v1/aircraft_types (not implemented) [intent=direct_read availability=not_implemented operation=aviationstack.get.v1-aircraft-types]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 cities - Documented GET /v1/cities (not implemented) [intent=direct_read availability=not_implemented operation=aviationstack.get.v1-cities]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 flightsfuture - Documented GET /v1/flightsFuture (not implemented) [intent=direct_read availability=not_implemented operation=aviationstack.get.v1-flightsfuture]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 historical - Documented GET /v1/historical (not implemented) [intent=direct_read availability=not_implemented operation=aviationstack.get.v1-historical]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 routes - Documented GET /v1/routes (not implemented) [intent=direct_read availability=not_implemented operation=aviationstack.get.v1-routes]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 taxes - Documented GET /v1/taxes (not implemented) [intent=direct_read availability=not_implemented operation=aviationstack.get.v1-taxes]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 timetable - Documented GET /v1/timetable (not implemented) [intent=direct_read availability=not_implemented operation=aviationstack.get.v1-timetable]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    countries list - Run the countries ETL stream [intent=etl availability=implemented stream=countries]
    flights list - Run the flights ETL stream [intent=etl availability=implemented stream=flights]

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
