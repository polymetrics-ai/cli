# Overview

Reads Nylas calendars, contacts, messages, and events for a connected grant through the Nylas v3
REST API.

Readable streams: `calendars`, `contacts`, `messages`, `events`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.nylas.com/docs/api/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Nylas v3 API key. Sent only as a Bearer token; never logged.
- `base_url` (optional, string); default `https://api.us.nylas.com`; format `uri`; Nylas API base
  URL. Defaults to the US region host; set to https://api.eu.nylas.com for the EU region.
- `calendar_id` (optional, string); Calendar id required by the events stream.
- `grant_id` (optional, string); default `me`; Nylas grant id to read. Defaults to 'me' (the
  connected grant).
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `50`; Records per page (1-200).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.us.nylas.com`, `grant_id=me`, `max_pages=0`,
`page_size=50`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v3/grants/{{ config.grant_id }}/calendars` with query `limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page_token`; next token from `next_cursor`.

- `calendars`: GET `/v3/grants/{{ config.grant_id }}/calendars` - records path `data`; query
  `limit`=`{{ config.page_size }}`; cursor pagination; cursor parameter `page_token`; next token
  from `next_cursor`.
- `contacts`: GET `/v3/grants/{{ config.grant_id }}/contacts` - records path `data`; query
  `limit`=`{{ config.page_size }}`; cursor pagination; cursor parameter `page_token`; next token
  from `next_cursor`.
- `messages`: GET `/v3/grants/{{ config.grant_id }}/messages` - records path `data`; query
  `limit`=`{{ config.page_size }}`; cursor pagination; cursor parameter `page_token`; next token
  from `next_cursor`.
- `events`: GET `/v3/grants/{{ config.grant_id }}/events` - records path `data`; query
  `calendar_id`=`{{ config.calendar_id }}`; `limit`=`{{ config.page_size }}`; cursor pagination;
  cursor parameter `page_token`; next token from `next_cursor`.

## Write actions & risks

This connector is read-only. Read behavior: external Nylas API read of a connected grant's calendar,
contact, and message data.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=5.
