# Overview

Reads Qonto bank transactions, memberships, and accounts through the Qonto REST API (read-only).

Readable streams: `transactions`, `memberships`, `accounts`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api-doc.qonto.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Qonto API key, typically formatted
  '<organization-slug>:<secret-key>'. Sent verbatim as the Authorization header (no Bearer prefix,
  matching Qonto's own convention). Never logged.
- `base_url` (optional, string); default `https://thirdparty.qonto.com/v2`; format `uri`; Qonto API
  base URL. Defaults to https://thirdparty.qonto.com/v2.
- `iban` (optional, string); IBAN of the Qonto bank account to read transactions for. Required for
  the transactions stream.
- `start_date` (optional, string); Lower bound sent as the start_date query parameter on every
  request, when set.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://thirdparty.qonto.com/v2`.

Authentication behavior:

- API key authentication in `Authorization` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/transactions` with query `iban`=`{{ config.iban }}`; `page`=`1`;
`per_page`=`1`.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `transactions`: GET `/transactions` - records path `transactions`; query `iban`=`{{ config.iban
  }}`; `page`=`1`; `per_page`=`100`; `start_date` from template `{{ config.start_date }}`, omitted
  when absent; cursor pagination; cursor parameter `page`; next token from `meta.next_page`;
  incremental cursor `settled_at`; formatted as `rfc3339`; computed output fields `id`.
- `memberships`: GET `/memberships` - records path `memberships`; query `page`=`1`;
  `per_page`=`100`; cursor pagination; cursor parameter `page`; next token from `meta.next_page`.
- `accounts`: GET `/accounts` - records path `accounts`; query `page`=`1`; `per_page`=`100`; cursor
  pagination; cursor parameter `page`; next token from `meta.next_page`.

## Write actions & risks

This connector is read-only. Read behavior: external Qonto API read of bank transaction and account
data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, non_data_endpoint=1, out_of_scope=3.
