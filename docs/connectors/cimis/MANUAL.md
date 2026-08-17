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
  mode
  api_key (secret)

ETL STREAMS
  stations:
    primary key: StationNbr
    fields: City(), ConnectDate(), County(), DisconnectDate(), Elevation(), GroundCover(), HmsLatitude(), HmsLongitude(), IsActive(), IsEtoStation(), Name(), RegionalOffice(), SitingDesc(), StationNbr(), ZipCodes()
  station_zip_codes:
    primary key: StationNbr, ZipCode
    fields: ConnectDate(), DisconnectDate(), IsActive(), StationNbr(), ZipCode()
  spatial_zip_codes:
    primary key: ZipCode
    fields: ConnectDate(), DisconnectDate(), IsActive(), ZipCode()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external CIMIS API read of public weather station metadata
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

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
