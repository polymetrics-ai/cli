# Overview

Reads Sentry projects, issues, error events, and releases through the Sentry REST API (read-only).

Readable streams: `projects`, `issues`, `events`, `releases`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.sentry.io/api/.

## Auth setup

Connection fields:

- `auth_token` (required, secret, string); Sentry internal integration / auth token. Used only for
  Bearer auth; never logged.
- `base_url` (required, string); format `uri`; Sentry API base URL, e.g. https://sentry.io or
  https://<self-hosted-host>.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `organization` (optional, string); Organization slug; required to read the issues, events, and
  releases streams.
- `page_size` (optional, string); default `100`; Records per page (1-100).
- `project` (optional, string); Project slug; required to read the issues and events streams.

Secret fields are redacted in logs and write previews: `auth_token`.

Default configuration values: `max_pages=0`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.auth_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/0/projects/`.

## Streams notes

Default pagination: single request; no pagination.

- `projects`: GET `/api/0/projects/` - records at response root.
- `issues`: GET `/api/0/projects/{{ config.organization }}/{{ config.project }}/issues/` - records
  at response root.
- `events`: GET `/api/0/projects/{{ config.organization }}/{{ config.project }}/events/` - records
  at response root.
- `releases`: GET `/api/0/organizations/{{ config.organization }}/releases/` - records at response
  root.

## Write actions & risks

This connector is read-only. Read behavior: external Sentry API read of project, issue, event, and
release data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=24, out_of_scope=180, requires_elevated_scope=13.
