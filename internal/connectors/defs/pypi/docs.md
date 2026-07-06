# Overview

Reads PyPI project metadata through the PyPI JSON API. Read-only and credential-free.

Readable streams: `project`.

This connector is read-only; no write actions are declared.

Service API documentation: https://warehouse.pypa.io/api-reference/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://pypi.org`; format `uri`; PyPI base URL. Defaults
  to https://pypi.org.
- `project_name` (required, string); PyPI project (package) name to read metadata for, e.g.
  'requests'. Must not contain '/', '?', '#', or '..'.

Default configuration values: `base_url=https://pypi.org`.

Authentication behavior:

- No authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/pypi/{{ config.project_name }}/json`.

## Streams notes

Default pagination: single request; no pagination.

- `project`: GET `/pypi/{{ config.project_name }}/json` - records path `info`; emits passthrough
  records.

## Write actions & risks

This connector is read-only. Read behavior: external PyPI JSON API read of public package metadata.

## Known limits

- Batch defaults: read_page_size=1.
- API coverage includes 1 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=2.
