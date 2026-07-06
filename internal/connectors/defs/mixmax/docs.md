# Overview

Reads Mixmax code snippets, messages, rules, sequences, and meeting types through the Mixmax REST
API.

Readable streams: `codesnippets`, `messages`, `rules`, `sequences`, `meetingtypes`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.mixmax.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Mixmax API token. Sent as the X-API-Token header. Never
  logged.
- `base_url` (optional, string); default `https://api.mixmax.com/v1`; format `uri`; Mixmax API base
  URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.mixmax.com/v1`.

Authentication behavior:

- API key authentication in `X-API-Token` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/codesnippets`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `next`; next token from `next`; stop flag
`hasNext`.

- `codesnippets`: GET `/codesnippets` - records path `results`; cursor pagination; cursor parameter
  `next`; next token from `next`; stop flag `hasNext`.
- `messages`: GET `/messages` - records path `results`; cursor pagination; cursor parameter `next`;
  next token from `next`; stop flag `hasNext`.
- `rules`: GET `/rules` - records path `results`; cursor pagination; cursor parameter `next`; next
  token from `next`; stop flag `hasNext`.
- `sequences`: GET `/sequences` - records path `results`; cursor pagination; cursor parameter
  `next`; next token from `next`; stop flag `hasNext`.
- `meetingtypes`: GET `/meetingtypes` - records path `results`; cursor pagination; cursor parameter
  `next`; next token from `next`; stop flag `hasNext`.

## Write actions & risks

This connector is read-only. Read behavior: external Mixmax API read of code snippet, message, rule,
sequence, and meeting-type data.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=3.
