# Overview

Reads Feishu/Lark Bitable (Base) records, tables, and field schemas via the Open Platform REST API
using a tenant_access_token exchange.

Readable streams: `records`, `tables`, `fields`.

This connector is read-only; no write actions are declared.

Service API documentation:
https://open.feishu.cn/document/server-docs/docs/bitable-v1/bitable-overview.

## Auth setup

Connection fields:

- `app_id` (required, secret, string); The unique identifier for your application. Found in the
  Feishu/Lark Developer Console under "Credentials & Basic Info".
- `app_secret` (required, secret, string); The secret key used to verify your application's
  identity. Found alongside the App ID.
- `app_token` (required, secret, string); The unique identifier of the Bitable (Base). Found in the
  URL: /base/{app_token}.
- `base_url` (optional, string).
- `lark_host` (required, string); Base URL of the Feishu/Lark Open Platform. Use
  https://open.feishu.cn for Feishu (China mainland) accounts and https://open.larksuite.com for
  Lark (international) accounts.
- `mode` (optional, string).
- `page_size` (optional, string); Number of records per request. Max: 500. Default: 100.
- `table_id` (required, string); The unique identifier of the table. Found in the URL query
  parameter table={table_id}.

Secret fields are redacted in logs and write previews: `app_id`, `app_secret`, `app_token`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

- `records`: GET connector-managed request path - records path `data`.
- `tables`: GET connector-managed request path - records path `data`.
- `fields`: GET connector-managed request path - records path `data`.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 3 stream-backed endpoint group(s).
