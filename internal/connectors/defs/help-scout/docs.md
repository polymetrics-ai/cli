# Overview

Reads Help Scout conversations, customers, mailboxes, and users through the Mailbox API using OAuth2
client-credentials authentication.

Readable streams: `conversations`, `customers`, `mailboxes`, `users`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.helpscout.com/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.helpscout.net/v2`; format `uri`; Help Scout
  Mailbox API base URL override for tests or proxies.
- `client_id` (required, secret, string); Help Scout OAuth2 application client id, used for the
  client-credentials token exchange. Never logged.
- `client_secret` (required, secret, string); Help Scout OAuth2 application client secret, used for
  the client-credentials token exchange. Never logged.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; sent as modifiedSince to
  scope conversations/customers to records changed at or after this time.
- `token_url` (optional, string); default `https://api.helpscout.net/v2/oauth2/token`; format `uri`;
  Help Scout OAuth2 token endpoint override for tests or proxies.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`.

Default configuration values: `base_url=https://api.helpscout.net/v2`,
`token_url=https://api.helpscout.net/v2/oauth2/token`.

Authentication behavior:

- OAuth 2.0 client credentials authentication using `config.token_url`, `secrets.client_id`,
  `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/mailboxes`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `size`; starts at
1; page size 50.

- `conversations`: GET `/conversations` - records path `_embedded.conversations`; query
  `modifiedSince` from template `{{ config.start_date }}`, omitted when absent;
  `sortField`=`modifiedAt`; `sortOrder`=`asc`; page-number pagination; page parameter `page`; size
  parameter `size`; starts at 1; page size 50.
- `customers`: GET `/customers` - records path `_embedded.customers`; query `modifiedSince` from
  template `{{ config.start_date }}`, omitted when absent; `sortField`=`modifiedAt`;
  `sortOrder`=`asc`; page-number pagination; page parameter `page`; size parameter `size`; starts at
  1; page size 50.
- `mailboxes`: GET `/mailboxes` - records path `_embedded.mailboxes`; query `modifiedSince` from
  template `{{ config.start_date }}`, omitted when absent; `sortField`=`modifiedAt`;
  `sortOrder`=`asc`; page-number pagination; page parameter `page`; size parameter `size`; starts at
  1; page size 50.
- `users`: GET `/users` - records path `_embedded.users`; query `modifiedSince` from template `{{
  config.start_date }}`, omitted when absent; `sortField`=`modifiedAt`; `sortOrder`=`asc`;
  page-number pagination; page parameter `page`; size parameter `size`; starts at 1; page size 50.

## Write actions & risks

This connector is read-only. Read behavior: external Help Scout API read of conversation, customer,
mailbox, and user data.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=3.
