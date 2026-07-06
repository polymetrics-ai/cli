# Overview

Reads California Irrigation Management Information System (CIMIS) weather station metadata and
station/spatial zip-code reference lists through the CIMIS Web API. Read-only.

Readable streams: `stations`, `station_zip_codes`, `spatial_zip_codes`.

This connector is read-only; no write actions are declared.

Service API documentation: https://cimis.water.ca.gov/WSNReportCriteria.aspx.

## Auth setup

Connection fields:

- `api_key` (optional, secret, string); CIMIS appKey. Not required for the stations stream (CIMIS
  serves station metadata without an appKey); reserved for a future daily/hourly stream expansion.
  Never logged.
- `base_url` (optional, string); default `https://et.water.ca.gov`; format `uri`; CIMIS API base URL
  override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://et.water.ca.gov`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/station`.

## Streams notes

Default pagination: single request; no pagination.

- `stations`: GET `/api/station` - records path `Stations`; emits passthrough records.
- `station_zip_codes`: GET `/api/stationzipcode` - records path `ZipCodes`; emits passthrough
  records.
- `spatial_zip_codes`: GET `/api/spatialzipcode` - records path `ZipCodes`; emits passthrough
  records.

## Write actions & risks

This connector is read-only. Read behavior: external CIMIS API read of public weather station
metadata.

## Known limits

- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=3, out_of_scope=1.
