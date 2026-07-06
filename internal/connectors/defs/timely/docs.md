# Overview

Reads users, projects, clients, calendar/time events, time entries (hours), tags (labels), and teams
from the Timely API. Read-only: every Timely mutation endpoint requires a nested single-key JSON
body envelope (e.g. {"client": {...}}) the connector cannot express.

Readable streams: `users`, `projects`, `clients`, `events`, `hours`, `labels`, `teams`.

This connector is read-only; no write actions are declared.

Service API documentation: https://dev.timelyapp.com/.

## Auth setup

Connection fields:

- `account_id` (required, string); Timely account id every stream's path is prefixed with
  (<account_id>/<resource>).
- `base_url` (optional, string); default `https://api.timelyapp.com/1.1`; format `uri`; Timely API
  base URL override for tests or proxies.
- `bearer_token` (required, secret, string); Timely OAuth access token, sent as a Bearer token
  (Authorization: Bearer <token>). Never logged.
- `start_date` (optional, string); Optional lower bound sent as the 'since' query param on the
  'events' stream only; omitted entirely (and every other stream unaffected) when unset.

Secret fields are redacted in logs and write previews: `bearer_token`.

Default configuration values: `base_url=https://api.timelyapp.com/1.1`.

Authentication behavior:

- Bearer token authentication using `secrets.bearer_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/{{ config.account_id }}/users`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `users`, `projects`, `clients`, `events`, `labels`, `teams`;
page_number: `hours`.

- `users`: GET `/{{ config.account_id }}/users` - records path `.`; emits passthrough records.
- `projects`: GET `/{{ config.account_id }}/projects` - records path `.`; emits passthrough records.
- `clients`: GET `/{{ config.account_id }}/clients` - records path `.`; emits passthrough records.
- `events`: GET `/{{ config.account_id }}/events` - records path `.`; query `since` from template
  `{{ config.start_date }}`, omitted when absent; emits passthrough records.
- `hours`: GET `/{{ config.account_id }}/hours` - records path `.`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; computed output fields
  `project_id`, `user_id`; emits passthrough records.
- `labels`: GET `/{{ config.account_id }}/labels` - records path `.`; emits passthrough records.
- `teams`: GET `/{{ config.account_id }}/teams` - records path `.`; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Timely API read of user, project, client, time
event/entry, tag, and team data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 7 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  deprecated=1, destructive_admin=5, duplicate_of=1, non_data_endpoint=4, out_of_scope=19,
  requires_elevated_scope=1.
