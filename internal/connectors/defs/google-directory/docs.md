# Overview

Reads Google Admin SDK Directory users, groups, organizational units, and ChromeOS devices via
bearer-token OAuth. Read-only.

Readable streams: `users`, `groups`, `orgunits`, `chromeos_devices`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.google.com/admin-sdk/directory/reference/rest.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Google OAuth 2.0 access token with Admin SDK Directory
  read scope. Used only for Bearer auth; never logged. Acquisition/refresh is out of scope for this
  connector (credentials layer already owns it) -- see docs.md Known limits.
- `base_url` (optional, string); default `https://admin.googleapis.com/admin/directory/v1`; format
  `uri`; Google Admin SDK Directory API base URL override for tests or proxies.
- `customer_id` (optional, string); default `my_customer`; Google Workspace customer id.
- `max_pages` (optional, string); Maximum pages fetched per stream. A positive integer, or
  'all'/'unlimited' (default) for no cap.
- `mode` (optional, string).
- `page_size` (optional, integer); default `100`; maxResults per page (1-500).

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://admin.googleapis.com/admin/directory/v1`,
`customer_id=my_customer`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `users` with query `customer`=`{{ config.customer_id }}`;
`maxResults`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `pageToken`; next token from
`nextPageToken`.

- `users`: GET `users` - records path `users`; query `customer`=`{{ config.customer_id }}`;
  `maxResults`=`{{ config.page_size }}`; cursor pagination; cursor parameter `pageToken`; next token
  from `nextPageToken`; computed output fields `id`, `name`, `org_unit_path`, `primary_email`.
- `groups`: GET `groups` - records path `groups`; query `customer`=`{{ config.customer_id }}`;
  `maxResults`=`{{ config.page_size }}`; cursor pagination; cursor parameter `pageToken`; next token
  from `nextPageToken`; computed output fields `description`, `email`, `id`, `name`.
- `orgunits`: GET `customer/{{ config.customer_id }}/orgunits` - records path `organizationUnits`;
  query `maxResults`=`{{ config.page_size }}`; cursor pagination; cursor parameter `pageToken`; next
  token from `nextPageToken`; computed output fields `description`, `id`, `name`, `org_unit_path`.
- `chromeos_devices`: GET `customer/{{ config.customer_id }}/devices/chromeos` - records path
  `chromeosdevices`; query `maxResults`=`{{ config.page_size }}`; cursor pagination; cursor
  parameter `pageToken`; next token from `nextPageToken`; computed output fields `id`,
  `org_unit_path`, `serial_number`, `status`.

## Write actions & risks

This connector is read-only. Read behavior: external Google Admin SDK Directory API read of
user/group/org-unit/device metadata.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=4.
