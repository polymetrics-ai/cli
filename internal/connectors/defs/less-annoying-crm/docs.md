# Overview

Reads Less Annoying CRM users, contacts, tasks, notes, and events through the Less Annoying CRM v2
API.

Readable streams: `users`, `contacts`, `tasks`, `notes`, `events`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.lessannoyingcrm.com/help/topic/API.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); API key to use. Manage and create your API keys on the
  Programmer API settings page at https://account.lessannoyingcrm.com/app/Settings/Api.
- `base_url` (optional, string).
- `mode` (optional, string).
- `start_date` (required, string).

Secret fields are redacted in logs and write previews: `api_key`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `users`: GET connector-managed request path - records path `data`.
- `contacts`: GET connector-managed request path - records path `data`.
- `tasks`: GET connector-managed request path - records path `data`; incremental cursor
  `DateCreated`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `notes`: GET connector-managed request path - records path `data`.
- `events`: GET connector-managed request path - records path `data`; incremental cursor
  `DateUpdated`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `tasks`, `events`.
