# Overview

Reads Genesys Cloud users, queues, groups, and divisions through the Genesys Cloud Platform API.

Readable streams: `users`, `queues`, `groups`, `divisions`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.genesys.cloud/devapps/api-explorer.

## Auth setup

Connection fields:

- `base_url` (required, string); format `uri`; Genesys Cloud Platform API base URL, e.g.
  https://api.mypurecloud.com/api/v2.
- `client_id` (required, secret, string); Genesys Cloud OAuth client credentials grant client id.
  Never logged.
- `client_secret` (required, secret, string); Genesys Cloud OAuth client credentials grant client
  secret. Never logged.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`.
- `scope` (optional, string); default ; Optional OAuth scope requested on the client-credentials
  token exchange.
- `token_url` (required, string); format `uri`; Genesys Cloud OAuth token endpoint, e.g.
  https://login.mypurecloud.com/oauth/token.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`.

Default configuration values: `max_pages=0`, `page_size=100`, `scope=`.

Authentication behavior:

- OAuth 2.0 client credentials authentication using `config.token_url`, `secrets.client_id`,
  `secrets.client_secret`, `config.scope`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `users` with query `pageNumber`=`1`; `pageSize`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `pageNumber`; size parameter `pageSize`;
starts at 1; page size 100.

- `users`: GET `users` - records path `entities`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; computed output fields
  `display_name`.
- `queues`: GET `routing/queues` - records path `entities`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100.
- `groups`: GET `groups` - records path `entities`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100.
- `divisions`: GET `authorization/divisions` - records path `entities`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external Genesys Cloud Platform API read of user, queue,
group, and division data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, non_data_endpoint=1, out_of_scope=3.
