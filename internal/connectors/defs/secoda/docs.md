# Overview

Reads Secoda catalog metadata (tables, documents, collections, questions) through the Secoda API.

Readable streams: `tables`, `documents`, `collections`, `questions`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.secoda.co/api.md.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Secoda API key, sent as a Bearer token (Authorization:
  Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.secoda.co/api/v1`; format `uri`; Secoda API
  base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.secoda.co/api/v1`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/tables`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `page_size`;
starts at 1; page size 100; maximum 1 page(s).

- `tables`: GET `/tables` - records path `results`; page-number pagination; page parameter `page`;
  size parameter `page_size`; starts at 1; page size 100; maximum 1 page(s); emits passthrough
  records.
- `documents`: GET `/documents` - records path `results`; page-number pagination; page parameter
  `page`; size parameter `page_size`; starts at 1; page size 100; maximum 1 page(s); emits
  passthrough records.
- `collections`: GET `/collections` - records path `results`; page-number pagination; page parameter
  `page`; size parameter `page_size`; starts at 1; page size 100; maximum 1 page(s); emits
  passthrough records.
- `questions`: GET `/questions` - records path `results`; page-number pagination; page parameter
  `page`; size parameter `page_size`; starts at 1; page size 100; maximum 1 page(s); emits
  passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Secoda API read of data-catalog metadata.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=4.
