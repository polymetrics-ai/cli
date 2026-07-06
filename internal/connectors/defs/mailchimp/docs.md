# Overview

Reads Mailchimp Marketing API audiences (lists), campaigns, reports, and automations through the
datacenter-scoped REST API.

Readable streams: `lists`, `campaigns`, `reports`, `automations`.

This connector is read-only; no write actions are declared.

Service API documentation: https://mailchimp.com/developer/release-notes/.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Mailchimp OAuth access token. Takes precedence over
  api_key when both are set. Sent as Bearer auth; never logged.
- `api_key` (optional, secret, string); Mailchimp API key (e.g. abc123-us6). Sent as HTTP Basic auth
  (username "anystring") when access_token is unset; never logged.
- `data_center` (required, string); Mailchimp datacenter token (e.g. "us6"), used to build the
  datacenter-scoped base URL https://<data_center>.api.mailchimp.com/3.0.
- `mode` (optional, string).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only objects
  created/sent at or after this time are read (per-stream since_* filter).

Secret fields are redacted in logs and write previews: `access_token`, `api_key`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token` when `{{ secrets.access_token }}`.
- HTTP Basic authentication using `secrets.api_key` when `{{ secrets.api_key }}`.

Requests use base URL `https://{{ config.data_center }}.api.mailchimp.com/3.0` after applying
configuration defaults.

Connection checks call GET `/lists` with query `count`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `count`;
page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `lists`: GET `/lists` - records path `lists`; offset/limit pagination; offset parameter `offset`;
  limit parameter `count`; page size 100; incremental cursor `date_created`; sent as
  `since_date_created`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `campaigns`: GET `/campaigns` - records path `campaigns`; offset/limit pagination; offset
  parameter `offset`; limit parameter `count`; page size 100; incremental cursor `create_time`; sent
  as `since_create_time`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `reports`: GET `/reports` - records path `reports`; offset/limit pagination; offset parameter
  `offset`; limit parameter `count`; page size 100; incremental cursor `send_time`; sent as
  `since_send_time`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `automations`: GET `/automations` - records path `automations`; offset/limit pagination; offset
  parameter `offset`; limit parameter `count`; page size 100; incremental cursor `create_time`; sent
  as `since_create_time`; formatted as `rfc3339`; initial lower bound from `start_date`.

## Write actions & risks

This connector is read-only. Read behavior: external Mailchimp API read of audience, campaign,
report, and automation data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5.
