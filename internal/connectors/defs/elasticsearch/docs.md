# Overview

Reads Elasticsearch index metadata and documents through the REST API. Read-only.

Readable streams: `indices`, `documents`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.elastic.co/docs/reference/elasticsearch.

## Auth setup

Connection fields:

- `api_key_id` (optional, secret, string); Elasticsearch API key id.
- `api_key_secret` (optional, secret, string); Elasticsearch API key secret. See api_key_id.
- `endpoint` (required, string); format `uri`; Elasticsearch cluster base URL (e.g.
  https://es.example.com:9200).
- `index` (optional, string); Index name to query for the 'documents' stream (required for that
  stream only).
- `max_pages` (optional, string); default `100`; Maximum pages to read for the 'documents' stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Documents per page for the 'documents' stream
  search (from/size pagination).
- `password` (optional, secret, string); Basic auth password.
- `username` (optional, string); Basic auth username, used when api_key_id/api_key_secret are not
  both set.

Secret fields are redacted in logs and write previews: `api_key_id`, `api_key_secret`, `password`.

Default configuration values: `max_pages=100`, `page_size=100`.

Authentication behavior:

- Connector-specific authentication using `config.api_key_id` when `{{ config.api_key_id }}`.
- HTTP Basic authentication using `config.username`, `secrets.password` when `{{ config.username
  }}`.
- No authentication.

Requests use the configured `endpoint` value after applying defaults.

Connection checks call GET `/`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `indices`; offset_limit: `documents`.

- `indices`: GET `/_cat/indices` - records at response root; query `format`=`json`.
- `documents`: GET `/{{ config.index }}/_search` - records path `hits.hits`; offset/limit
  pagination; offset parameter `from`; limit parameter `size`; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external Elasticsearch cluster read of index metadata
and documents.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, non_data_endpoint=1, out_of_scope=1.
