# Overview

Reads Care Quality Commission (CQC) registered locations, providers, and inspection areas from the
public CQC Syndication API. Read-only (full-refresh).

Readable streams: `locations`, `providers`, `inspection_areas`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.cqc.org.uk/about-us/transparency/using-cqc-data.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); CQC Syndication API primary subscription key, sent as the
  Ocp-Apim-Subscription-Key header. Never logged.
- `base_url` (optional, string); default `https://api.service.cqc.org.uk/public/v1`; format `uri`;
  CQC Syndication API base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.service.cqc.org.uk/public/v1`.

Authentication behavior:

- API key authentication in `Ocp-Apim-Subscription-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/inspection-areas`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `inspection_areas`; page_number: `locations`, `providers`.

- `locations`: GET `/locations` - records path `locations`; page-number pagination; page parameter
  `page`; size parameter `perPage`; starts at 1; page size 1000.
- `providers`: GET `/providers` - records path `providers`; page-number pagination; page parameter
  `page`; size parameter `perPage`; starts at 1; page size 1000.
- `inspection_areas`: GET `/inspection-areas` - records path `inspectionAreas`.

## Write actions & risks

This connector is read-only. Read behavior: external CQC Syndication API read of publicly published
care provider/location data.

## Known limits

- Batch defaults: read_page_size=1000.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=2, out_of_scope=3.
