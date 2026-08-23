---
name: pm-cimis
description: CIMIS connector knowledge and safe action guide.
---

# pm-cimis

## Purpose

Reads California Irrigation Management Information System (CIMIS) weather station metadata and station/spatial zip-code reference lists through the CIMIS Web API. Read-only.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- api_key (secret)

## ETL Streams

- stations:
  - primary key: StationNbr
  - fields: City(string), ConnectDate(string), County(string), DisconnectDate(string), Elevation(string), GroundCover(string), HmsLatitude(string), HmsLongitude(string), IsActive(string), IsEtoStation(string), Name(string), RegionalOffice(string), SitingDesc(string), StationNbr(string), ZipCodes(array)
- station_zip_codes:
  - primary key: StationNbr, ZipCode
  - fields: ConnectDate(string), DisconnectDate(string), IsActive(string), StationNbr(integer), ZipCode(string)
- spatial_zip_codes:
  - primary key: ZipCode
  - fields: ConnectDate(string), DisconnectDate(string), IsActive(string), ZipCode(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external CIMIS API read of public weather station metadata
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect cimis
```

### Inspect as structured JSON

```bash
pm connectors inspect cimis --json
```

## Agent Rules

- Run pm connectors inspect cimis before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
