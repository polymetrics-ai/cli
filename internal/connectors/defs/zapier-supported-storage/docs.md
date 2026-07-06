# Overview

Reads and writes Zapier Storage key/value records.

Readable streams: `records`.

Write actions: `set_record`, `increment_record`, `delete_record`, `delete_all_records`.

Service API documentation:
https://help.zapier.com/hc/en-us/articles/8496293271053-Save-and-retrieve-data-from-Zaps.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://store.zapier.com`; format `uri`; Zapier Storage
  API root. Also usable as a base URL override for tests/proxies.
- `mode` (optional, string).
- `secret` (required, secret, string); Zapier Storage secret, sent as the 'secret' query parameter
  on every request. Never logged.

Secret fields are redacted in logs and write previews: `secret`.

Default configuration values: `base_url=https://store.zapier.com`.

Authentication behavior:

- API key authentication in query parameter `secret` using `secrets.secret`.

Requests use the configured `base_url` value after applying defaults.

## Streams notes

Default pagination: single request; no pagination.

- `records`: GET `/api/records` - records path `records`.

## Write actions & risks

Overall write risk: external mutation of a shared per-Zap/per-app key/value store: set/increment a
single key, delete a single key, or wipe the entire bucket (delete_all_records, destructive).

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `set_record`: PATCH `/api/records` - kind `create`; body type `json`; required record fields
  `action`, `data`; accepted fields `action`, `data`; risk: creates or overwrites a single key/value
  pair in the caller's Zapier Storage bucket (optionally only when the existing value matches
  only_if_value); external mutation, no approval required.
- `increment_record`: PATCH `/api/records` - kind `update`; body type `json`; required record fields
  `action`, `data`; accepted fields `action`, `data`; risk: atomically increments a numeric-valued
  key by amount (creating it at amount if absent); external mutation, no approval required.
- `delete_record`: DELETE `/api/records?key={{ record.key }}` - kind `delete`; body type `none`;
  path fields `key`; required record fields `key`; accepted fields `key`; missing records treated as
  success for status `404`; risk: irreversibly deletes a single key from the caller's Zapier Storage
  bucket.
- `delete_all_records`: DELETE `/api/records` - kind `delete`; body type `none`; confirmation
  `destructive`; risk: irreversibly deletes EVERY key in the caller's Zapier Storage bucket
  (whole-bucket wipe); destructive, requires explicit confirmation.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 1 stream-backed endpoint group(s), 4 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=2, out_of_scope=5.
