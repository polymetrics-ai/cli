# Overview

Reads RingCentral extensions, call logs, messages, contacts, and devices through the REST API.

Readable streams: `extensions`, `call_log`, `messages`, `contacts`, `devices`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.ringcentral.com/api-reference.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); RingCentral OAuth2 access token, sent as an
  Authorization: Bearer <access_token> header. Never logged.
- `base_url` (optional, string); default `https://platform.ringcentral.com/restapi/v1.0`; format
  `uri`; RingCentral API base URL override for tests or proxies.
- `dateFrom` (optional, string); Optional passthrough filter: only return call-log/message records
  at or after this value.
- `dateTo` (optional, string); Optional passthrough filter: only return call-log/message records at
  or before this value.
- `direction` (optional, string); Optional passthrough filter: scope the call-log/messages streams
  to a single direction (Inbound/Outbound).
- `messageType` (optional, string); Optional passthrough filter: scope the messages stream to a
  single message type.
- `type` (optional, string); Optional passthrough filter: scope the call-log stream to a single call
  type.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://platform.ringcentral.com/restapi/v1.0`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/account/~`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `perPage`; starts
at 1; page size 100.

- `extensions`: GET `/account/~/extension` - records path `records`; query `dateFrom` from template
  `{{ config.dateFrom }}`, omitted when absent; `dateTo` from template `{{ config.dateTo }}`,
  omitted when absent; `direction` from template `{{ config.direction }}`, omitted when absent;
  `messageType` from template `{{ config.messageType }}`, omitted when absent; `type` from template
  `{{ config.type }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `perPage`; starts at 1; page size 100; computed output fields `extension_number`,
  `stream`; emits passthrough records.
- `call_log`: GET `/account/~/extension/~/call-log` - records path `records`; query `dateFrom` from
  template `{{ config.dateFrom }}`, omitted when absent; `dateTo` from template `{{ config.dateTo
  }}`, omitted when absent; `direction` from template `{{ config.direction }}`, omitted when absent;
  `messageType` from template `{{ config.messageType }}`, omitted when absent; `type` from template
  `{{ config.type }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `perPage`; starts at 1; page size 100; computed output fields `start_time`, `stream`;
  emits passthrough records.
- `messages`: GET `/account/~/extension/~/message-store` - records path `records`; query `dateFrom`
  from template `{{ config.dateFrom }}`, omitted when absent; `dateTo` from template `{{
  config.dateTo }}`, omitted when absent; `direction` from template `{{ config.direction }}`,
  omitted when absent; `messageType` from template `{{ config.messageType }}`, omitted when absent;
  `type` from template `{{ config.type }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `perPage`; starts at 1; page size 100; computed output fields
  `creation_time`, `stream`; emits passthrough records.
- `contacts`: GET `/account/~/extension/~/address-book/contact` - records path `records`; query
  `dateFrom` from template `{{ config.dateFrom }}`, omitted when absent; `dateTo` from template `{{
  config.dateTo }}`, omitted when absent; `direction` from template `{{ config.direction }}`,
  omitted when absent; `messageType` from template `{{ config.messageType }}`, omitted when absent;
  `type` from template `{{ config.type }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `perPage`; starts at 1; page size 100; computed output fields
  `first_name`, `last_name`, `stream`; emits passthrough records.
- `devices`: GET `/account/~/device` - records path `records`; query `dateFrom` from template `{{
  config.dateFrom }}`, omitted when absent; `dateTo` from template `{{ config.dateTo }}`, omitted
  when absent; `direction` from template `{{ config.direction }}`, omitted when absent;
  `messageType` from template `{{ config.messageType }}`, omitted when absent; `type` from template
  `{{ config.type }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `perPage`; starts at 1; page size 100; computed output fields `stream`; emits
  passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external RingCentral API read of account extension,
call-log, message, contact, and device data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=5.
