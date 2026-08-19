# Overview

Reads and writes the documented Help Scout Mailbox API v2 surface using OAuth2 client-credentials
authentication.

Current official operation ledger: 144 documented HTTP operations (79 GET, 21 POST, 20 PUT, 18
DELETE, 6 PATCH). Implemented rows: 139 = 24 stream-backed reads + 49 bounded direct reads + 65
typed writes + 1 binary download. Blocked/planned rows: 5. Certified rows: 0 (fixture-only; no live
provider calls were made).

Readable streams: `conversations`, `conversations_threads`, `customer_properties`, `customers`,
`customers_chats`, `customers_emails`, `customers_phones`, `customers_social_profiles`,
`customers_websites`, `mailboxes`, `mailboxes_fields`, `mailboxes_folders`,
`mailboxes_saved_replies`, `organizations`, `organizations_properties`,
`organizations_conversations`, `organizations_customers`, `tags`, `teams`, `teams_members`, `users`,
`users_status`, `webhooks`, `workflows`.

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

- `conversationid`, `customerid`, `mailboxid`, `organizationid`, `teamid` (all optional, string);
  parent ids that scope the twelve sub-resource streams. Supply them per read with the generated
  `--conversation-id`, `--customer-id`, `--mailbox-id`, `--organization-id`, and `--team-id` flags
  rather than storing them on the connection.

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

Twelve sub-resource streams are scoped by a parent id supplied per read rather than fanned out over
every parent. Their paths interpolate a connection config value, and the generated ETL command
carries the matching required flag: `conversations_threads` uses `config.conversationid`
(`--conversation-id`); `customers_chats`, `customers_emails`, `customers_phones`,
`customers_social_profiles` and `customers_websites` use `config.customerid` (`--customer-id`);
`mailboxes_fields`, `mailboxes_folders` and `mailboxes_saved_replies` use `config.mailboxid`
(`--mailbox-id`); `organizations_conversations` and `organizations_customers` use
`config.organizationid` (`--organization-id`); `teams_members` uses `config.teamid` (`--team-id`).

## Write actions & risks

The connector declares 65 typed write actions (21 POST creates, 20 PUT and 6 PATCH updates, 18
DELETE removals) across conversations, threads, customers and their email/phone/chat/social/website
contact records, organizations, users, teams, tags, webhooks, and workflows.

Writes are only available through reverse ETL plan -> preview -> explicit approval -> execute. The
18 DELETE actions are gated as destructive and additionally require a typed confirmation; they
permanently remove Help Scout records, including `delete_customer`, `delete_conversation`,
`delete_attachment`, `delete_email`, and `delete_phone`. The bundle does not expose arbitrary
request bodies, raw query strings, generic method/path/body, file bytes, shell commands, or
passthrough HTTP tools.

Seven create actions (`create_conversation`, `create_customer`, `create_customer_property_definition`,
`create_organization`, `create_organization_property_definition`, `create_user`, `create_webhook`)
declare an open `record_schema`: the extraction artifact carried no closed request body for them, so
their required provider fields are not enforced by this bundle. Closing those bodies is a named
dependency on a Help Scout request-body specification; until it lands, a plan built for those seven
is validated by the provider rather than by preflight.

Read behavior: external Help Scout API read of conversation, thread, customer, organization,
mailbox, team, tag, webhook, workflow, and user data.

## Known limits

- Batch defaults: read_page_size=50.
- `sync_transport.json` declares the 24 current fixture streams and
  `conversations(id) → update_conversation(conversationId)` as a typed-destination proof. It has
  keyed delivery, durable acknowledgement, all closed modes, and fixture/preflight evidence against
  the persisted App/CLI typed-destination foundation at
  `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57`. Provider-live certification remains pending.
- The source inventory remains unproven. Every declared typed action has an explicit eligibility
  disposition. `update_customer` is not reverse-ETL enabled while the shared
  `declarative-typed-destination-action-specific-source-bindings` foundation gap remains open: the
  closed contract currently allows only one source binding for the declarative source executor, so
  its required `customers(id) → customerId` mapping cannot be added beside the conversation mapping
  without a connector-specific workaround. Other typed actions are declaration-pending their exact
  source mapping and conformance evidence. Direct-read gaps and source-inventory recovery remain
  connector work, not a safety or certification exclusion.
- Fixture-only evidence: no live Help Scout credentials, provider calls, provider writes, or
  certification run were used.
