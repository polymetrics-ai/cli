# pm connectors inspect cimis

```text
NAME
  pm connectors inspect cimis - CIMIS connector manual

SYNOPSIS
  pm connectors inspect cimis
  pm connectors inspect cimis --json
  pm credentials add <name> --connector cimis [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads California Irrigation Management Information System (CIMIS) weather station metadata and station/spatial zip-code reference lists through the CIMIS Web API. Read-only.

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
  api_key (secret)

ETL STREAMS
  stations:
    primary key: StationNbr
    fields: City(string), ConnectDate(string), County(string), DisconnectDate(string), Elevation(string), GroundCover(string), HmsLatitude(string), HmsLongitude(string), IsActive(string), IsEtoStation(string), Name(string), RegionalOffice(string), SitingDesc(string), StationNbr(string), ZipCodes(array)
  station_zip_codes:
    primary key: StationNbr, ZipCode
    fields: ConnectDate(string), DisconnectDate(string), IsActive(string), StationNbr(integer), ZipCode(string)
  spatial_zip_codes:
    primary key: ZipCode
    fields: ConnectDate(string), DisconnectDate(string), IsActive(string), ZipCode(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external CIMIS API read of public weather station metadata
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run CIMIS's declared streams and reverse-ETL actions.
  Usage: pm cimis <command> [flags]
  Read streams
  Other Commands
    api get api data - Documented GET /api/data (not implemented) [intent=direct_read availability=not_implemented operation=cimis.get.api-data]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api spatialzipcode zipcode - Documented GET /api/spatialzipcode/{zipCode} (not implemented) [intent=direct_read availability=not_implemented operation=cimis.get.api-spatialzipcode-zipcode]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api station stationnumber - Documented GET /api/station/{stationNumber} (not implemented) [intent=direct_read availability=not_implemented operation=cimis.get.api-station-stationnumber]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api stationzipcode zipcode - Documented GET /api/stationzipcode/{zipCode} (not implemented) [intent=direct_read availability=not_implemented operation=cimis.get.api-stationzipcode-zipcode]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    spatial zip codes list - Run the spatial zip codes ETL stream [intent=etl availability=implemented stream=spatial_zip_codes]
    station zip codes list - Run the station zip codes ETL stream [intent=etl availability=implemented stream=station_zip_codes]
    stations list - Run the stations ETL stream [intent=etl availability=implemented stream=stations]

EXAMPLES
  # Inspect as a manual
  pm connectors inspect cimis

  # Inspect as structured JSON
  pm connectors inspect cimis --json

AGENT WORKFLOW
  - Run pm connectors inspect cimis before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
