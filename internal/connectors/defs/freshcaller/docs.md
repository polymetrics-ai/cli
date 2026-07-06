# Overview

Reads Freshcaller calls, agents, teams, and phone numbers through the Freshcaller REST API.

Readable streams: `calls`, `agents`, `teams`, `numbers`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.freshcaller.com/api/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Freshcaller API key, sent as the Authorization header in the
  form 'Token token=<api_key>'. Never logged.
- `base_url` (optional, string); default `https://api.freshcaller.com/api/v1`; format `uri`;
  Freshcaller API base URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.freshcaller.com/api/v1`, `max_pages=0`,
`page_size=100`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Token token=` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `calls` with query `page`=`1`; `page_size`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `page_size`;
starts at 1; page size 100.

- `calls`: GET `calls` - records path `calls`; page-number pagination; page parameter `page`; size
  parameter `page_size`; starts at 1; page size 100.
- `agents`: GET `agents` - records path `agents`; page-number pagination; page parameter `page`;
  size parameter `page_size`; starts at 1; page size 100.
- `teams`: GET `teams` - records path `teams`; page-number pagination; page parameter `page`; size
  parameter `page_size`; starts at 1; page size 100.
- `numbers`: GET `numbers` - records path `numbers`; page-number pagination; page parameter `page`;
  size parameter `page_size`; starts at 1; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external Freshcaller API read of call, agent, team, and
phone number data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=3.
