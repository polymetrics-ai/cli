# Overview

Reads records from a configured Infor Nexus export dataset through the Infor Nexus Data API (v3.1)
using HMAC-SHA256 request signing. Read-only.

Readable streams: `datasets`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.infor.com/inforos/nexus/.

## Auth setup

Connection fields:

- `access_key_id` (required, secret, string); Infor Nexus HMAC access key id. Sent verbatim on the
  X-Infor-AccessKeyId header and as the credential prefix of the signed Authorization header; never
  logged.
- `api_key` (required, secret, string); Infor Nexus Data API key. Sent verbatim on the
  X-Infor-ApiKey header; never logged.
- `base_url` (required, string); format `uri`; Infor Nexus Data API (v3.1) base URL, e.g.
  https://mingle-portal.inforcloudsuite.com/TENANT/ns/api/v3.1.
- `dataset_name` (required, string); Name of the configured Infor Nexus export dataset to read
  (path-escaped into /datasets/{dataset_name}).
- `mode` (optional, string).
- `secret_key` (required, secret, string); Infor Nexus HMAC secret key. Used only to compute the
  HMAC-SHA256 request signature (never sent on the wire itself, never logged).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound for the modifiedSince
  query filter on the first sync (superseded by the persisted incremental cursor on subsequent
  syncs).
- `user_id` (required, secret, string); Infor Nexus user id. Sent verbatim on the X-Infor-UserId
  header; never logged.

Secret fields are redacted in logs and write previews: `access_key_id`, `api_key`, `secret_key`,
`user_id`.

Authentication behavior:

- Connector-specific authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/datasets/{{ config.dataset_name }}` with query `limit`=`1`;
`offset`=`0`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `datasets`: GET `/datasets/{{ config.dataset_name }}` - records path `records`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100; incremental cursor
  `updated_at`; sent as `modifiedSince`; formatted as `rfc3339`; initial lower bound from
  `start_date`; computed output fields `dataset_name`, `id`, `updated_at`.

## Write actions & risks

This connector is read-only. Read behavior: external Infor Nexus dataset export read, HMAC-signed.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 1 stream-backed endpoint group(s).
