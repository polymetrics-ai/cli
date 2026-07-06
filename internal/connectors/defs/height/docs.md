# Overview

Reads Height tasks, lists, field templates, users, and workspace through the Height REST API.

Readable streams: `tasks`, `lists`, `field_templates`, `users`, `workspace`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.height.app/api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Height API key, sent as the Authorization header
  (Authorization: api-key <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.height.app`; format `uri`; Height API base URL
  override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.height.app`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `api-key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/workspace`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `tasks`; none: `lists`, `field_templates`, `users`, `workspace`.

- `tasks`: GET `/tasks` - records path `list`; query `usePagination`=`true`; cursor pagination;
  cursor parameter `after`; next token from `nextPageToken`.
- `lists`: GET `/lists` - records path `list`.
- `field_templates`: GET `/fieldTemplates` - records path `list`.
- `users`: GET `/users` - records path `list`.
- `workspace`: GET `/workspace` - records at response root.

## Write actions & risks

This connector is read-only. Read behavior: external Height API read of task, list, field-template,
user, and workspace data.

## Known limits

- Batch defaults: read_page_size=1000.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=3.
