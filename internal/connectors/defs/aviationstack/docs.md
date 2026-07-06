# Overview

Reads aviationstack flights and aviation reference data (airlines, airports, airplanes, countries)
through the aviationstack REST API. Read-only.

Readable streams: `flights`, `airlines`, `airports`, `airplanes`, `countries`.

This connector is read-only; no write actions are declared.

Service API documentation: https://aviationstack.com/documentation.

## Auth setup

Connection fields:

- `access_key` (required, secret, string); Aviationstack API access key. Sent as the access_key
  query parameter; never logged.
- `base_url` (optional, string); default `https://api.aviationstack.com/v1`; format `uri`;
  Aviationstack API base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `access_key`.

Default configuration values: `base_url=https://api.aviationstack.com/v1`.

Authentication behavior:

- API key authentication in query parameter `access_key` using `secrets.access_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/countries` with query `limit`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `flights`: GET `/flights` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; incremental cursor `flight_date`; formatted as
  `rfc3339`; computed output fields `airline_iata`, `airline_name`, `arrival_airport`,
  `arrival_iata`, `arrival_scheduled`, `departure_airport`, `departure_iata`, `departure_scheduled`,
  `flight_iata`, `flight_icao`, `flight_number`.
- `airlines`: GET `/airlines` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `airports`: GET `/airports` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `airplanes`: GET `/airplanes` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `countries`: GET `/countries` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external aviationstack API read of flight and aviation
reference data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3, requires_elevated_scope=4.
