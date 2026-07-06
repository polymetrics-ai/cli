# Overview

Reads Marketo leads, programs, and activities through Marketo REST endpoints. Read-only; does not
refresh OAuth tokens internally.

Readable streams: `leads`, `programs`, `activities`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.marketo.com/rest-api/.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Marketo REST API access token, sent as a Bearer token.
  This connector does not refresh OAuth tokens internally; the caller must supply a valid, unexpired
  token.
- `activity_type_ids` (optional, string); Optional comma-separated activityTypeIds filter, sent as
  the activityTypeIds query parameter on the activities stream only.
- `base_url` (required, string); format `uri`; Your Marketo REST identity host base URL, ending in
  /rest/v1 (tenant-specific; no default). Example: https://123-ABC-456.mktorest.com/rest/v1.
- `max_pages` (optional, string); default `0`; Maximum pages to read; use 0, all, or unlimited to
  exhaust the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `300`; Records per page (1-300), sent as the batchSize
  query parameter.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `max_pages=0`, `page_size=300`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/leads.json` with query `batchSize`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `nextPageToken`; next token from
`nextPageToken`; stop flag `moreResult`.

- `leads`: GET `/leads.json` - records path `result`; query `batchSize`=`{{ config.page_size }}`;
  cursor pagination; cursor parameter `nextPageToken`; next token from `nextPageToken`; stop flag
  `moreResult`; computed output fields `createdAt`, `email`, `id`, `updatedAt`.
- `programs`: GET `/programs.json` - records path `result`; query `batchSize`=`{{ config.page_size
  }}`; cursor pagination; cursor parameter `nextPageToken`; next token from `nextPageToken`; stop
  flag `moreResult`; computed output fields `createdAt`, `id`, `name`, `updatedAt`.
- `activities`: GET `/activities.json` - records path `result`; query `activityTypeIds` from
  template `{{ config.activity_type_ids }}`, omitted when absent; `batchSize`=`{{ config.page_size
  }}`; cursor pagination; cursor parameter `nextPageToken`; next token from `nextPageToken`; stop
  flag `moreResult`; computed output fields `activityDate`, `activityTypeId`, `id`, `leadId`.

## Write actions & risks

This connector is read-only. Read behavior: external Marketo REST API read of lead, program, and
activity data.

## Known limits

- Batch defaults: read_page_size=300.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=3.
