# Overview

Reads Mailtrap accounts, inboxes, projects, and sending domains through the Mailtrap
account-management REST API.

Readable streams: `accounts`, `inboxes`, `projects`, `sending_domains`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api-docs.mailtrap.io/.

## Auth setup

Connection fields:

- `account_id` (optional, string); Mailtrap account id; required for account-scoped streams
  (inboxes, projects, sending_domains), substituted into the account-scoped path.
- `api_token` (required, secret, string); Mailtrap API token, sent as a Bearer token (Authorization:
  Bearer <api_token>). Never logged.
- `base_url` (optional, string); default `https://mailtrap.io/api`; format `uri`; Mailtrap API base
  URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://mailtrap.io/api`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/accounts`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

- `accounts`: GET `/accounts` - records path `.`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100.
- `inboxes`: GET `/accounts/{{ config.account_id }}/inboxes` - records path `.`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; computed
  output fields `account_id`.
- `projects`: GET `/accounts/{{ config.account_id }}/projects` - records path `.`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; computed
  output fields `account_id`.
- `sending_domains`: GET `/accounts/{{ config.account_id }}/sending_domains` - records path `data`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `account_id`.

## Write actions & risks

This connector is read-only. Read behavior: external Mailtrap API read of account-management data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=3.
