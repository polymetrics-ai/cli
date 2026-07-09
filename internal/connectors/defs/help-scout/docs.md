# Overview

Reads Help Scout conversations, customers, mailboxes, and users through the Inbox API using OAuth2
client-credentials authentication.

Readable streams: `conversations`, `customers`, `mailboxes`, `users`.

Service API documentation: https://developer.helpscout.com/mailbox-api/.

This bundle now tracks the full official Inbox API endpoint navigation in `api_surface.json`: 146
docs endpoint pages, 145 unique method/path rows, and method split GET 79, POST 21, PUT 20, PATCH 6,
DELETE 19. Only the four stream-backed reads are executable today; all other documented operations
are blocked-by-default metadata for the follow-up direct-read, binary, reverse-ETL, and
sensitive/admin policy lanes.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.helpscout.net/v2`; format `uri`; Help Scout
  Inbox API base URL override for tests or proxies.
- `client_id` (required, secret, string); Help Scout OAuth2 application client id, used for the
  client-credentials token exchange. Never logged.
- `client_secret` (required, secret, string); Help Scout OAuth2 application client secret, used for
  the client-credentials token exchange. Never logged.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; sent as `modifiedSince`
  to scope conversations/customers/mailboxes/users to records changed at or after this time.
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

This connector is read-only at runtime. No `writes.json` actions are declared, and no Help Scout
mutation is executable from this bundle today.

Documented mutation endpoints are recorded in `api_surface.json` as blocked-by-default operation
rows instead of blanket exclusions. Future write work must use explicit reverse-ETL actions with
record schemas, path fields, risk text, preview output, approval requirements, and destructive typed
confirmation where required. The bundle must not expose a raw generic HTTP write, shell write, SQL
write, or raw mutation escape hatch.

## Known limits

- Batch defaults: read_page_size=50.
- Runtime coverage is limited to 4 stream-backed endpoint groups.
- `api_surface.json` is full-surface metadata for the official Inbox API docs navigation, not a claim
  that every row is executable.
- Direct-read candidates such as single conversation/customer/mailbox/user lookups, tags, teams,
  workflows, webhooks, reports, and organization/customer property reads are blocked until #217/#218
  add bounded command and output policies.
- Binary or raw-message endpoints such as attachment downloads and thread original source are
  blocked until a bounded binary policy exists.
- Sensitive/admin/destructive mutations are blocked until #219 adds typed confirmation, redaction,
  preflight, and approval policy metadata.
- The official docs contain two pages for `GET /v2/conversations/{conversation_id}/threads/{thread_id}/original-source`
  (JSON and RFC 822 variants); the surface ledger tracks one method/path row and notes the duplicate.
