# Overview

Reads public Whisky Hunter auction and distillery data. Read-only, no credentials required.

Readable streams: `auctions`, `distilleries`, `auctions_data`, `auctions_info`, `distilleries_info`,
`auction_data`, `distillery_data`.

This connector is read-only; no write actions are declared.

Service API documentation: https://whiskyhunter.net/api/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://whiskyhunter.net`; format `uri`; Whisky Hunter API
  base URL override for tests or proxies.

Default configuration values: `base_url=https://whiskyhunter.net`.

Authentication behavior:

- No authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/auctions_data/`.

## Streams notes

Default pagination: single request; no pagination.

- `auctions`: GET `/api/auctions_data/` - records path `.`; emits passthrough records.
- `distilleries`: GET `/api/distilleries_info/` - records path `.`; emits passthrough records.
- `auctions_data`: GET `/api/auctions_data/` - records path `.`; emits passthrough records.
- `auctions_info`: GET `/api/auctions_info` - records path `.`; emits passthrough records.
- `distilleries_info`: GET `/api/distilleries_info/` - records path `.`; emits passthrough records.
- `auction_data`: GET `/api/auction_data/{{ fanout.id }}/` - records path `.`; fan-out; ids from
  request `/api/auctions_info`; id-list records path `.`; id field `slug`; id inserted into the
  request path; emits passthrough records.
- `distillery_data`: GET `/api/distillery_data/{{ fanout.id }}/` - records path `.`; fan-out; ids
  from request `/api/distilleries_info/`; id-list records path `.`; id field `slug`; id inserted
  into the request path; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Whisky Hunter API read of public auction and
distillery data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 7 stream-backed endpoint group(s).
