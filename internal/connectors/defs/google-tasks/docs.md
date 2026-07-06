# Overview

Reads Google task lists and tasks through the Google Tasks REST API.

Readable streams: `tasklists`, `tasks`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.google.com/tasks/reference/rest.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Google Tasks OAuth access token. Used only for Bearer auth;
  never logged.
- `base_url` (optional, string); default `https://tasks.googleapis.com/tasks/v1`; format `uri`;
  Google Tasks API base URL override for tests or proxies.
- `mode` (optional, string).
- `records_limit` (optional, integer); default `50`; Records requested per page (maxResults query
  param). Google Tasks caps this at 100.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://tasks.googleapis.com/tasks/v1`, `records_limit=50`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users/@me/lists` with query `maxResults`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `pageToken`; next token from
`nextPageToken`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `tasklists`: GET `/users/@me/lists` - records path `items`; query `maxResults`=`{{
  config.records_limit }}`; cursor pagination; cursor parameter `pageToken`; next token from
  `nextPageToken`; incremental cursor `updated`; formatted as `rfc3339`; computed output fields
  `self_link`.
- `tasks`: GET `/lists/{{ fanout.id }}/tasks` - records path `items`; query `maxResults`=`{{
  config.records_limit }}`; cursor pagination; cursor parameter `pageToken`; next token from
  `nextPageToken`; incremental cursor `updated`; formatted as `rfc3339`; computed output fields
  `self_link`; fan-out; ids from request `/users/@me/lists`; id-list records path `items`; id field
  `id`; id inserted into the request path; stamps `tasklist_id`.

## Write actions & risks

This connector is read-only. Read behavior: external Google Tasks API read of the authenticated
user's task lists and tasks.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 2 stream-backed endpoint group(s).
