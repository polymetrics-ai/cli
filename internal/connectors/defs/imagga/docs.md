# Overview

Reads Imagga account API usage and per-image tags/categories via the Imagga REST API. Read-only. The
colors and faces_detections detection streams are not yet ported - see docs.md Known limits.

Readable streams: `usage`, `tags`, `categories`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.imagga.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Imagga API key, sent as the HTTP Basic auth username. Never
  logged.
- `api_secret` (required, secret, string); Imagga API secret, sent as the HTTP Basic auth password.
  Never logged.
- `base_url` (optional, string); default `https://api.imagga.com/v2`; format `uri`; Imagga API base
  URL override for tests or proxies.
- `image_urls` (optional, string); default
  `https://imagga.com/static/images/categorization/child-476506_640.jpg`; Comma-separated image URLs
  to analyze via the tags/categories/colors/faces_detections streams.

Secret fields are redacted in logs and write previews: `api_key`, `api_secret`.

Default configuration values: `base_url=https://api.imagga.com/v2`,
`image_urls=https://imagga.com/static/images/categorization/child-476506_640.jpg`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`, `secrets.api_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/usage`.

## Streams notes

Default pagination: single request; no pagination.

- `usage`: GET `/usage` - single-object response; records path `result`; computed output fields
  `daily_processed`, `monthly_limit`, `monthly_processed`, `period`, `requests`.
- `tags`: GET `/tags` - records path `result.tags`; computed output fields `confidence`, `tag`;
  fan-out; ids from config field `image_urls`; id sent as query parameter `image_url`; stamps
  `image_url`.
- `categories`: GET `/categories/personal_photos` - records path `result.categories`; computed
  output fields `category`, `confidence`; fan-out; ids from config field `image_urls`; id sent as
  query parameter `image_url`; stamps `image_url`.

## Write actions & risks

This connector is read-only. Read behavior: external Imagga API read of account usage data and
per-image tags/categories.

## Known limits

- Batch defaults: read_page_size=1.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=2.
