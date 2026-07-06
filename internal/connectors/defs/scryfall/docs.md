# Overview

Reads cards and sets from the public Scryfall API. Read-only and credential-free.

Readable streams: `cards`, `sets`.

This connector is read-only; no write actions are declared.

Service API documentation: https://scryfall.com/docs/api.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.scryfall.com`; format `uri`; Scryfall API base
  URL override for tests or proxies.
- `q` (optional, string); default `*`; Scryfall search query for the 'cards' stream (Scryfall search
  syntax).

Default configuration values: `base_url=https://api.scryfall.com`, `q=*`.

Authentication is handled by the connector-specific implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/cards/search`.

## Streams notes

Default pagination: single request; no pagination.

- `cards`: GET `/cards/search` - records path `data`; query `q`=`{{ config.q }}`; follows a
  next-page URL from the response body; URL path `next_page`; next URLs stay on the configured API
  host; emits passthrough records.
- `sets`: GET `/sets` - records path `data`; follows a next-page URL from the response body; URL
  path `next_page`; next URLs stay on the configured API host; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: public, credential-free Scryfall API read of card and
set data.

## Known limits

- API coverage includes 2 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=3.
