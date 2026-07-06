# Overview

Zendesk Sunshine reads 4 stream(s), and writes through 7 action(s).

Readable streams: `object_types`, `relationship_types`, `objects`, `relationships`.

Write actions: `create_object_record`, `upsert_object_record_by_external_id`,
`delete_object_record`, `delete_object_type`, `create_relationship_record`,
`delete_relationship_record`, `delete_relationship_type`.

Service API documentation:
https://developer.zendesk.com/api-reference/custom-data/custom-objects-api/introduction/.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); Zendesk API token, used as the Basic auth password
  alongside '{email}/token' as the username.
- `base_url` (required, string); format `uri`; Your Zendesk Sunshine API base URL (e.g.
  https://yoursubdomain.zendesk.com/api/sunshine). No derived default: must be provided explicitly
  (see docs.md Known limits).
- `email` (required, string); Zendesk agent email address, combined with api_token for HTTP Basic
  auth as '{email}/token'.
- `object_type` (optional, string).
- `relationship_type` (optional, string); Same 400-without-a-filter caveat as object_type, see
  docs.md Known limits.

Secret fields are redacted in logs and write previews: `api_token`.

Authentication behavior:

- HTTP Basic authentication using `config.email`, `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/objects/types`.

## Streams notes

Default pagination: single request; no pagination.

- `object_types`: GET `/objects/types` - records path `data`; computed output fields `id`.
- `relationship_types`: GET `/relationships/types` - records path `data`; computed output fields
  `id`.
- `objects`: GET `/objects/records` - records path `data`; query `type` from template `{{
  config.object_type }}`, omitted when absent; computed output fields `id`.
- `relationships`: GET `/relationships/records` - records path `data`; query `type` from template
  `{{ config.relationship_type }}`, omitted when absent; computed output fields `id`.

## Write actions & risks

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_object_record`: POST `/objects/records` - kind `create`; body type `json`; required record
  fields `data`; accepted fields `data`.
- `upsert_object_record_by_external_id`: PUT `/objects/records` - kind `upsert`; body type `json`;
  required record fields `data`; accepted fields `data`; risk: creates a new record with the given
  external_id if none exists, or overwrites the attributes object of the existing record with that
  external_id and type -- an overwrite, not a merge: any attribute omitted from this call's
  attributes is cleared on an existing record.
- `delete_object_record`: DELETE `/objects/records/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`.
- `delete_object_type`: DELETE `/objects/types/{{ record.key }}` - kind `delete`; body type `none`;
  path fields `key`; required record fields `key`; accepted fields `key`; missing records treated as
  success for status `404`; risk: Standard Zendesk object types (users, tickets, organizations)
  cannot be deleted this way and the API rejects the request.
- `create_relationship_record`: POST `/relationships/records` - kind `create`; body type `json`;
  required record fields `data`; accepted fields `data`.
- `delete_relationship_record`: DELETE `/relationships/records/{{ record.id }}` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; risk: permanently removes a relationship record
  between two object records; irreversible, does not affect either underlying record.
- `delete_relationship_type`: DELETE `/relationships/types/{{ record.key }}` - kind `delete`; body
  type `none`; path fields `key`; required record fields `key`; accepted fields `key`; missing
  records treated as success for status `404`.

## Known limits

- API coverage includes 4 stream-backed endpoint group(s), 7 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  deprecated=3, duplicate_of=3, out_of_scope=3.
