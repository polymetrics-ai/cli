# Overview

Reads Lokalise project keys, languages, translations, contributors, and comments through fixed API v2 project routes.

Readable streams: `keys`, `languages`, `translations`, `contributors`, `comments`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.lokalise.com/reference/api-overview.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); read-access API token sent in the declared `X-Api-Token` header.
- `project_id` (required, string); Lokalise project identifier.

Secret fields are redacted in logs and write previews: `api_key`.

The runtime uses only the fixed `https://api.lokalise.com/api2` origin. It does not accept caller-provided origins or fixture modes.

Connection checks read one language through the declared route.

## Streams notes

Default pagination: numbered `page` requests with `limit=100` and a fixed ten-thousand-page safety cap.

Incremental streams use their declared cursor fields and client-side lower-bound filtering when a lower bound is available.

- `keys`: GET `/projects/{project_id}/keys`; incremental cursor `modified_at_timestamp`.
- `languages`: GET `/projects/{project_id}/languages`.
- `translations`: GET `/projects/{project_id}/translations`; incremental cursor `modified_at_timestamp`.
- `contributors`: GET `/projects/{project_id}/contributors`.
- `comments`: GET `/projects/{project_id}/comments`.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 5 stream-backed endpoint groups.
