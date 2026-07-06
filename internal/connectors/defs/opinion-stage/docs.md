# Overview

Reads Opinion Stage items (polls, quizzes, and forms) through the Opinion Stage Public Result API.
Read-only.

Readable streams: `items`, `responses`, `questions`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.opinionstage.com/api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Opinion Stage personal API key, from
  https://app.opinionstage.com/dashboard/settings/account/edit. Sent as the HTTP Basic auth username
  with an empty password (the Public Result API's documented API-key auth convention); never logged.
- `base_url` (optional, string); default `https://api.opinionstage.com`; format `uri`; Opinion Stage
  Public Result API base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.opinionstage.com`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/v2/items`.

## Streams notes

Default pagination: page-number pagination; page parameter `page[number]`; size parameter
`page[size]`; starts at 1; page size 50.

- `items`: GET `/api/v2/items` - records path `data`; page-number pagination; page parameter
  `page[number]`; size parameter `page[size]`; starts at 1; page size 50; computed output fields
  `created`, `embed`, `modified`, `status`, `title`.
- `responses`: GET `/api/v2/items/{{ fanout.id }}/responses` - records path `data`; page-number
  pagination; page parameter `page[number]`; size parameter `page[size]`; starts at 1; page size 50;
  computed output fields `answers`, `created`, `duration`, `result`, `result_text`, `result_title`,
  `utm`; fan-out; ids from request `/api/v2/items`; id-list records path `data`; id field `id`; id
  inserted into the request path; stamps `item_id`.
- `questions`: GET `/api/v2/items/{{ fanout.id }}/questions` - records path `data`; page-number
  pagination; page parameter `page[number]`; size parameter `page[size]`; starts at 1; page size 50;
  computed output fields `created`, `kind`, `lead`, `modified`, `title`; fan-out; ids from request
  `/api/v2/items`; id-list records path `data`; id field `id`; id inserted into the request path;
  stamps `item_id`.

## Write actions & risks

This connector is read-only. Read behavior: external Opinion Stage API read of item directory.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 3 stream-backed endpoint group(s).
