# Overview

Reads Mailjet contacts, contact lists, messages, campaigns, and statistics through the Mailjet Email
REST API (v3).

Readable streams: `contacts`, `contactslists`, `messages`, `campaigns`, `stats`.

This connector is read-only; no write actions are declared.

Service API documentation: https://dev.mailjet.com/email/reference/.

## Auth setup

Connection fields:

- `api_key` (required, string); Mailjet public API key, sent as the username of HTTP Basic auth.
- `api_key_secret` (required, secret, string); Mailjet secret API key, sent as the password of HTTP
  Basic auth. Never logged.
- `base_url` (optional, string); default `https://api.mailjet.com/v3/REST`; format `uri`; Mailjet
  Email REST API (v3) base URL override for tests or proxies.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-1000).

Secret fields are redacted in logs and write previews: `api_key_secret`.

Default configuration values: `base_url=https://api.mailjet.com/v3/REST`, `page_size=100`.

Authentication behavior:

- HTTP Basic authentication using `config.api_key`, `secrets.api_key_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/contact`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `Offset`; limit parameter `Limit`;
page size 100.

- `contacts`: GET `/contact` - records path `Data`; offset/limit pagination; offset parameter
  `Offset`; limit parameter `Limit`; page size 100.
- `contactslists`: GET `/contactslist` - records path `Data`; offset/limit pagination; offset
  parameter `Offset`; limit parameter `Limit`; page size 100.
- `messages`: GET `/message` - records path `Data`; offset/limit pagination; offset parameter
  `Offset`; limit parameter `Limit`; page size 100.
- `campaigns`: GET `/campaign` - records path `Data`; offset/limit pagination; offset parameter
  `Offset`; limit parameter `Limit`; page size 100.
- `stats`: GET `/statcounters` - records path `Data`; offset/limit pagination; offset parameter
  `Offset`; limit parameter `Limit`; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external Mailjet API read of contact, list, message,
campaign, and statistics data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=6.
