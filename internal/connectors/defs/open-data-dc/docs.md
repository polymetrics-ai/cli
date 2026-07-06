# Overview

Reads District of Columbia Master Address Repository (MAR 2) locations, units, and SSL parcel
records via the Open Data DC API. Read-only.

Readable streams: `locations`, `units`, `ssls`.

This connector is read-only; no write actions are declared.

Service API documentation: https://opendata.dc.gov/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Open Data DC (MAR 2) API key, sent as the apikey query
  parameter. Never logged.
- `base_url` (optional, string); default `https://datagate.dc.gov/mar/open/api/v2.2`; format `uri`;
  Open Data DC API base URL override for tests or proxies.
- `location` (optional, string); Address, place, or block search term for the 'locations' stream.
  Required for that stream only.
- `marid` (optional, string); MAR id for the 'units' stream (required for that stream) and an
  optional filter for the 'ssls' stream.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://datagate.dc.gov/mar/open/api/v2.2`.

Authentication behavior:

- API key authentication in query parameter `apikey` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/ssls`.

## Streams notes

Default pagination: single request; no pagination.

- `locations`: GET `/locations/{{ config.location }}` - records path `Result.addresses`; computed
  output fields `AddrNum`, `Anc`, `CensusTract`, `FullAddress`, `Latitude`, `Longitude`, `MarId`,
  `Quadrant`, `ResidenceType`, `SSL`, `StName`, `Status`, `Ward`, `Xcoord`, `Ycoord`, and 2 more.
- `units`: GET `/units/{{ config.marid }}` - records path `Result.units`.
- `ssls`: GET `/ssls` - records path `Result.ssls`; query `marid` from template `{{ config.marid
  }}`, omitted when absent.

## Write actions & risks

This connector is read-only. Read behavior: external Open Data DC (MAR 2) API read of public
address/parcel data.

## Known limits

- API coverage includes 3 stream-backed endpoint group(s).
