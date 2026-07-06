# Overview

Reads Mux Video assets, live streams, direct uploads, and system signing keys through the Mux REST
API using HTTP Basic authentication.

Readable streams: `assets`, `live_streams`, `uploads`, `signing_keys`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.mux.com/api-reference.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.mux.com`; format `uri`; Mux API base URL
  override for tests or proxies.
- `mode` (optional, string).
- `password` (required, secret, string); Mux API access token secret, used as the HTTP Basic auth
  password. Never logged.
- `username` (required, string); Mux API access token id, used as the HTTP Basic auth username.

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `base_url=https://api.mux.com`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/video/v1/assets`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 25.

- `assets`: GET `/video/v1/assets` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 25.
- `live_streams`: GET `/video/v1/live-streams` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 25.
- `uploads`: GET `/video/v1/uploads` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 25.
- `signing_keys`: GET `/system/v1/signing-keys` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 25.

## Write actions & risks

This connector is read-only. Read behavior: external Mux API read of video asset, live stream,
upload, and signing key data.

## Known limits

- Batch defaults: read_page_size=25.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=5.
