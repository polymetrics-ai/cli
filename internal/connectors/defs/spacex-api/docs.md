# Overview

Reads public SpaceX launch, rocket, core, capsule, crew, Dragon, history, payload, and Starlink
data.

Readable streams: `launches`, `rockets`, `capsules`, `cores`, `crew`, `dragons`, `history`,
`payloads`, `starlink`, `launchpads`, `landpads`, `ships`, `roadster`, `company`.

This connector is read-only; no write actions are declared.

Service API documentation: https://github.com/r-spacex/SpaceX-API/tree/master/docs.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.spacexdata.com/v4`; format `uri`; SpaceX API
  base URL override for tests or proxies.
- `mode` (optional, string).

Default configuration values: `base_url=https://api.spacexdata.com/v4`.

Authentication behavior:

- No authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/company`.

## Streams notes

Default pagination: single request; no pagination.

- `launches`: GET `/launches` - records path `.`; emits passthrough records.
- `rockets`: GET `/rockets` - records path `.`; emits passthrough records.
- `capsules`: GET `/capsules` - records path `.`; emits passthrough records.
- `cores`: GET `/cores` - records path `.`; emits passthrough records.
- `crew`: GET `/crew` - records path `.`; emits passthrough records.
- `dragons`: GET `/dragons` - records path `.`; emits passthrough records.
- `history`: GET `/history` - records path `.`; emits passthrough records.
- `payloads`: GET `/payloads` - records path `.`; emits passthrough records.
- `starlink`: GET `/starlink` - records path `.`; emits passthrough records.
- `launchpads`: GET `/launchpads` - records path `.`; emits passthrough records.
- `landpads`: GET `/landpads` - records path `.`; emits passthrough records.
- `ships`: GET `/ships` - records path `.`; emits passthrough records.
- `roadster`: GET `/roadster` - single-object response; records path `.`; emits passthrough records.
- `company`: GET `/company` - single-object response; records path `.`; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external public SpaceX API read of launch and vehicle
data.

## Known limits

- API coverage includes 14 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=17, out_of_scope=14.
