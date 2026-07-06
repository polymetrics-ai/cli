# Overview

Reads Kyriba bank accounts, transactions, statements, and payments through tenant REST API
collection endpoints. Read-only.

Readable streams: `bank_accounts`, `transactions`, `statements`, `payments`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.kyriba.com/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.kyriba.com/api/v1`; format `uri`; Kyriba
  tenant REST API base URL override. Kyriba deployments vary by tenant; defaults to
  https://api.kyriba.com/api/v1.
- `client_id` (required, secret, string); Kyriba OAuth2 client-credentials client id.
- `client_secret` (required, secret, string); Kyriba OAuth2 client-credentials client secret. Never
  logged.
- `scope` (optional, string); default ; Optional OAuth2 scope requested at token exchange. Left
  unset, no scope is sent (Kyriba tenant defaults apply).
- `token_url` (optional, string); default `https://api.kyriba.com/oauth/token`; format `uri`; Kyriba
  OAuth2 client-credentials token endpoint override. Defaults to https://api.kyriba.com/oauth/token.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`.

Default configuration values: `base_url=https://api.kyriba.com/api/v1`, `scope=`,
`token_url=https://api.kyriba.com/oauth/token`.

Authentication behavior:

- OAuth 2.0 client credentials authentication using `config.token_url`, `secrets.client_id`,
  `secrets.client_secret`, `config.scope`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/bank-accounts` with query `page`=`1`; `size`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `size`; starts at
1; page size 100.

- `bank_accounts`: GET `/bank-accounts` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `size`; starts at 1; page size 100; computed output fields
  `account_number`.
- `transactions`: GET `/transactions` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `size`; starts at 1; page size 100; computed output fields
  `account_number`.
- `statements`: GET `/statements` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `size`; starts at 1; page size 100; computed output fields
  `account_number`.
- `payments`: GET `/payments` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `size`; starts at 1; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external Kyriba tenant REST API read of bank
accounts/transactions/statements/payments.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=2.
