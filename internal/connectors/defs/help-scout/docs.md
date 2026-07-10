# Overview

Reads Help Scout conversations, customers, mailboxes, and users through the Inbox API using OAuth2
client-credentials authentication. The connector also exposes bounded JSON direct-read commands and
typed reverse-ETL write actions for the official Inbox API operation surface.

Readable streams: `conversations`, `customers`, `mailboxes`, `users`.

Operation inventory: 145 unique official method/path rows from the Help Scout Inbox API docs
navigation (GET 79, POST 21, PUT 20, PATCH 6, DELETE 19).

Runtime coverage in this bundle:

- 4 stream-backed ETL reads.
- 73 bounded JSON direct-read commands using `output_policy=json` and a 1 MiB response cap.
- 66 typed reverse-ETL write actions with explicit path fields, record schemas, risk text, approval
  requirements, and typed destructive confirmation.
- 2 binary/raw payload endpoints represented by bounded `binary_download` operation metadata; these
  remain feature-gated until a safe binary executor is enabled.

Service API documentation: https://developer.helpscout.com/mailbox-api/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.helpscout.net`; format `uri`; Help Scout API
  origin override for tests or proxies. The bundle sends versioned `/v2` and `/v3` paths explicitly.
- `client_id` (required, secret, string); Help Scout OAuth2 application client id, used for the
  client-credentials token exchange. Never logged.
- `client_secret` (required, secret, string); Help Scout OAuth2 application client secret, used for
  the client-credentials token exchange. Never logged.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; sent as `modifiedSince`
  to scope conversations/customers/mailboxes/users to records changed at or after this time.
- `token_url` (optional, string); default `https://api.helpscout.net/v2/oauth2/token`; format `uri`;
  Help Scout OAuth2 token endpoint override for tests or proxies.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`.

Default configuration values: `base_url=https://api.helpscout.net`,
`token_url=https://api.helpscout.net/v2/oauth2/token`.

Authentication behavior:

- OAuth 2.0 client credentials authentication using `config.token_url`, `secrets.client_id`,
  `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v2/mailboxes`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `size`; starts at
1; page size 50.

- `conversations`: GET `/v2/conversations` - records path `_embedded.conversations`; query
  `modifiedSince` from template `{{ config.start_date }}`, omitted when absent;
  `sortField`=`modifiedAt`; `sortOrder`=`asc`; page-number pagination; page parameter `page`; size
  parameter `size`; starts at 1; page size 50.
- `customers`: GET `/v2/customers` - records path `_embedded.customers`; query `modifiedSince` from
  template `{{ config.start_date }}`, omitted when absent; `sortField`=`modifiedAt`;
  `sortOrder`=`asc`; page-number pagination; page parameter `page`; size parameter `size`; starts at
  1; page size 50.
- `mailboxes`: GET `/v2/mailboxes` - records path `_embedded.mailboxes`; query `modifiedSince` from
  template `{{ config.start_date }}`, omitted when absent; `sortField`=`modifiedAt`;
  `sortOrder`=`asc`; page-number pagination; page parameter `page`; size parameter `size`; starts at
  1; page size 50.
- `users`: GET `/v2/users` - records path `_embedded.users`; query `modifiedSince` from template
  `{{ config.start_date }}`, omitted when absent; `sortField`=`modifiedAt`; `sortOrder`=`asc`;
  page-number pagination; page parameter `page`; size parameter `size`; starts at 1; page size 50.

## Write actions & risks

Help Scout mutations are available only as typed reverse-ETL actions. Connector-specific command
writes create a local plan first; execution still requires preview, approval token, and
`--confirm destructive`. No command exposes an arbitrary method/path/body escape hatch.

Examples:

```bash
pm help-scout conversations reply create --connection help-scout-prod --conversation-id 123 --text "Thanks" --preview --json
pm help-scout customers overwrite --connection help-scout-prod --customer-id 456 --firstName Ada --preview --json
pm help-scout webhooks create --connection help-scout-prod --url https://example.invalid/webhook --preview --json
```

Safety behavior:

- All 66 non-binary mutation endpoints are declared in `writes.json` and mapped from
  `api_surface.json` via `covered_by.write`.
- Every write action declares path fields and a record schema with `additionalProperties=false`.
- All write commands declare risk and approval text.
- All write actions use `confirm: destructive`, so approved execution requires the typed
  confirmation gate in addition to the approval token.
- Delete actions are idempotent for 404 responses where the engine can safely treat missing targets
  as already absent.

Binary/raw payload endpoints are not write actions and are not executed as unbounded file downloads.
They are represented as bounded `binary_download` operations with `max_bytes=104857600` and remain
feature-gated by the operation executor.

## Known limits

- Batch defaults: read_page_size=50, write_batch_size=1.
- Direct reads require saved Help Scout credentials and return JSON only; binary/raw message payloads
  remain feature-gated operation metadata.
- Some Help Scout write schemas are derived from official request examples. Fields not shown in the
  docs examples are intentionally not accepted by connector-specific command plans; use follow-up
  schema refinement rather than adding raw payload passthroughs.
- Complex object/array request fields can be supplied through reverse-ETL records. Connector command
  flags cover path parameters and scalar/example fields only.
- No credentialed Help Scout checks were run while authoring this bundle.
- The official docs contain two pages for `GET /v2/conversations/{conversation_id}/threads/{thread_id}/original-source`
  (JSON and RFC 822 variants); the surface ledger tracks one method/path row and represents it as a
  bounded binary/raw payload operation.
