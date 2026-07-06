# Overview

Reads Zendesk Chat agents, chats, departments, shortcuts, and triggers through the Zendesk Chat REST
API v2.

Readable streams: `agents`, `chats`, `departments`, `shortcuts`, `triggers`.

This connector is read-only; no write actions are declared.

Service API documentation:
https://developer.zendesk.com/api-reference/live-chat/chat-api/introduction/.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Zendesk Chat OAuth access token. Sent as Authorization:
  Bearer <access_token>; never logged.
- `base_url` (required, string); format `uri`; Your Zendesk Chat account root, e.g.
  https://acme.zendesk.com/api/v2/chat for subdomain 'acme'. Also usable as a base URL override for
  tests/proxies.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound for the chats
  incremental-export stream; converted to a Unix-seconds start_time query value.

Secret fields are redacted in logs and write previews: `access_token`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token` when `{{ secrets.access_token }}`.
- No authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/agents`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: next_url: `chats`; none: `agents`, `departments`, `shortcuts`, `triggers`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `agents`: GET `/agents` - records path `.`.
- `chats`: GET `/chats` - records path `chats`; follows a next-page URL from the response body; URL
  path `next_url`; next URLs stay on the configured API host; incremental cursor `timestamp`; sent
  as `start_time`; formatted as Unix-seconds timestamp; initial lower bound from `start_date`.
- `departments`: GET `/departments` - records path `.`.
- `shortcuts`: GET `/shortcuts` - records path `.`.
- `triggers`: GET `/triggers` - records path `.`.

## Write actions & risks

This connector is read-only. Read behavior: external Zendesk Chat API read of agent, chat, and
configuration data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=5.
