# Overview

Reads documents, folders, groups, and annotations from fixed Mendeley REST routes.

Readable streams: `documents`, `folders`, `groups`, `annotations`.

This connector is read-only; no write actions are declared.

Service API documentation: https://dev.mendeley.com/reference.

## Auth setup

Connection fields:

- `client_id` (required, secret, string); Mendeley OAuth client identifier.
- `client_refresh_token` (required, secret, string); refresh token for the declared OAuth grant.
- `client_secret` (required, secret, string); Mendeley OAuth client secret.
- `name_for_institution`, `query_for_catalog`, `start_date` (required, string); retained declared connection configuration.

Secret fields are redacted in logs and write previews: `client_id`, `client_refresh_token`, `client_secret`.

The runtime exchanges the refresh token only with the fixed Mendeley token endpoint and reads only the fixed `https://api.mendeley.com` origin. It does not accept caller-provided origins or fixture modes.

Connection checks make one bounded documents request.

## Streams notes

Default pagination: Link-header navigation with `limit=100`.

Incremental streams use their declared cursor fields and client-side lower-bound filtering when a lower bound is available.

- `documents`: GET `/documents`; `Accept: application/vnd.mendeley-document.1+json`; incremental cursor `last_modified`.
- `folders`: GET `/folders`; `Accept: application/vnd.mendeley-folder.1+json`; incremental cursor `modified`.
- `groups`: GET `/groups`; `Accept: application/vnd.mendeley-group.1+json`.
- `annotations`: GET `/annotations`; `Accept: application/vnd.mendeley-annotation.1+json`; incremental cursor `last_modified`.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 4 stream-backed endpoint groups.
