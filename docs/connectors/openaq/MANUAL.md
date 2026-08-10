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
  api_key (secret)

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
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external OpenAQ API read of public air-quality reference data
  approval: none; read-only public reference API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run OpenAQ's declared streams and reverse-ETL actions.
  Usage: pm openaq <command> [flags]
  Read streams
  Other Commands
    api get v3 countries countries-id - Documented GET /v3/countries/{countries_id} (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-countries-countries-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 instruments instruments-id - Documented GET /v3/instruments/{instruments_id} (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-instruments-instruments-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 licenses - Documented GET /v3/licenses (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-licenses]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 licenses licenses-id - Documented GET /v3/licenses/{licenses_id} (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-licenses-licenses-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 locations id sensors - Documented GET /v3/locations/{id}/sensors (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-locations-id-sensors]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 locations locations-id - Documented GET /v3/locations/{locations_id} (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-locations-locations-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 locations locations-id flags - Documented GET /v3/locations/{locations_id}/flags (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-locations-locations-id-flags]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 locations locations-id latest - Documented GET /v3/locations/{locations_id}/latest (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-locations-locations-id-latest]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 locations locations-id sensors - Documented GET /v3/locations/{locations_id}/sensors (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-locations-locations-id-sensors]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 manufacturers manufacturers-id - Documented GET /v3/manufacturers/{manufacturers_id} (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-manufacturers-manufacturers-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 manufacturers manufacturers-id instruments - Documented GET /v3/manufacturers/{manufacturers_id}/instruments (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-manufacturers-manufacturers-id-instruments]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 owners - Documented GET /v3/owners (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-owners]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 owners owners-id - Documented GET /v3/owners/{owners_id} (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-owners-owners-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 parameters parameters-id - Documented GET /v3/parameters/{parameters_id} (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-parameters-parameters-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 parameters parameters-id latest - Documented GET /v3/parameters/{parameters_id}/latest (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-parameters-parameters-id-latest]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 providers - Documented GET /v3/providers (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-providers]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 providers providers-id - Documented GET /v3/providers/{providers_id} (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-providers-providers-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 sensors id measurements - Documented GET /v3/sensors/{id}/measurements (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-sensors-id-measurements]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 sensors sensor-id flags - Documented GET /v3/sensors/{sensor_id}/flags (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-sensors-sensor-id-flags]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 sensors sensors-id - Documented GET /v3/sensors/{sensors_id} (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-sensors-sensors-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 sensors sensors-id days - Documented GET /v3/sensors/{sensors_id}/days (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-sensors-sensors-id-days]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 sensors sensors-id days dayofweek - Documented GET /v3/sensors/{sensors_id}/days/dayofweek (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-sensors-sensors-id-days-dayofweek]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 sensors sensors-id days monthly - Documented GET /v3/sensors/{sensors_id}/days/monthly (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-sensors-sensors-id-days-monthly]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 sensors sensors-id days monthofyear - Documented GET /v3/sensors/{sensors_id}/days/monthofyear (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-sensors-sensors-id-days-monthofyear]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 sensors sensors-id days yearly - Documented GET /v3/sensors/{sensors_id}/days/yearly (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-sensors-sensors-id-days-yearly]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 sensors sensors-id hours - Documented GET /v3/sensors/{sensors_id}/hours (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-sensors-sensors-id-hours]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 sensors sensors-id hours daily - Documented GET /v3/sensors/{sensors_id}/hours/daily (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-sensors-sensors-id-hours-daily]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 sensors sensors-id hours dayofweek - Documented GET /v3/sensors/{sensors_id}/hours/dayofweek (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-sensors-sensors-id-hours-dayofweek]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 sensors sensors-id hours hourofday - Documented GET /v3/sensors/{sensors_id}/hours/hourofday (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-sensors-sensors-id-hours-hourofday]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 sensors sensors-id hours monthly - Documented GET /v3/sensors/{sensors_id}/hours/monthly (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-sensors-sensors-id-hours-monthly]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 sensors sensors-id hours monthofyear - Documented GET /v3/sensors/{sensors_id}/hours/monthofyear (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-sensors-sensors-id-hours-monthofyear]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 sensors sensors-id hours yearly - Documented GET /v3/sensors/{sensors_id}/hours/yearly (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-sensors-sensors-id-hours-yearly]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 sensors sensors-id measurements - Documented GET /v3/sensors/{sensors_id}/measurements (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-sensors-sensors-id-measurements]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 sensors sensors-id measurements daily - Documented GET /v3/sensors/{sensors_id}/measurements/daily (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-sensors-sensors-id-measurements-daily]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 sensors sensors-id measurements hourly - Documented GET /v3/sensors/{sensors_id}/measurements/hourly (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-sensors-sensors-id-measurements-hourly]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 sensors sensors-id years - Documented GET /v3/sensors/{sensors_id}/years (not implemented) [intent=direct_read availability=not_implemented operation=openaq.get.v3-sensors-sensors-id-years]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    countries list - Run the countries ETL stream [intent=etl availability=implemented stream=countries]
    instruments list - Run the instruments ETL stream [intent=etl availability=implemented stream=instruments]
    locations list - Run the locations ETL stream [intent=etl availability=implemented stream=locations]
    manufacturers list - Run the manufacturers ETL stream [intent=etl availability=implemented stream=manufacturers]
    parameters list - Run the parameters ETL stream [intent=etl availability=implemented stream=parameters]

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
