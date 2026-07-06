# Overview

Reads Drift users, accounts, conversations, contacts, and teams, and writes
contact/account/message/conversation/timeline-event/GDPR mutations through the Drift REST API.

Readable streams: `users`, `accounts`, `conversations`, `contacts`, `teams`.

Write actions: `create_contact`, `update_contact`, `delete_contact`, `post_timeline_event`,
`create_account`, `update_account`, `delete_account`, `create_message`, `create_conversation`,
`gdpr_retrieve`, `gdpr_delete`.

Service API documentation: https://devdocs.drift.com/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Drift OAuth access token. Used only for Bearer auth;
  never logged.
- `base_url` (optional, string); default `https://driftapi.com`; format `uri`; Drift API base URL
  override for tests or proxies.
- `email` (optional, string); Optional email filter for the contacts stream lookup.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://driftapi.com`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users/list`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `conversations`; next_url: `accounts`; none: `users`, `contacts`,
`teams`.

- `users`: GET `/users/list` - records path `data`.
- `accounts`: GET `/accounts` - records path `data.accounts`; query `index`=`0`; `size`=`65`;
  follows a next-page URL from the response body; URL path `data.next`; next URLs stay on the
  configured API host; computed output fields `account_id`.
- `conversations`: GET `/conversations/list` - records path `data`; query `limit`=`50`; cursor
  pagination; cursor parameter `next`; next token from `pagination.next`; stop flag
  `pagination.more`.
- `contacts`: GET `/contacts` - records path `data`; query `email` from template `{{ config.email
  }}`, omitted when absent.
- `teams`: GET `/teams/org` - records path `data`.

## Write actions & risks

Overall write risk: external Drift API mutation of contacts, accounts, conversations, messages,
timeline events, and GDPR data-subject requests; delete_contact/delete_account/gdpr_delete are
destructive and require approval.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_contact`: POST `/contacts` - kind `create`; body type `json`; required record fields
  `attributes`; accepted fields `attributes`; risk: creates a new Drift contact record; low-risk
  external mutation, no approval required.
- `update_contact`: PATCH `/contacts/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `attributes`; accepted fields `attributes`, `id`; risk: mutates
  an existing Drift contact's attributes, including standard fields (email/name/phone) and any
  custom attribute; external mutation, approval required.
- `delete_contact`: DELETE `/contacts/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: permanently removes a Drift contact and its
  conversation history association; destructive, approval required.
- `post_timeline_event`: POST `/contacts/timeline` - kind `create`; body type `json`; required
  record fields `contactId`, `event`; accepted fields `attributes`, `contactId`, `createdAt`,
  `event`, `externalId`; risk: posts a custom timeline event onto a contact's record; low-risk
  external mutation, no approval required.
- `create_account`: POST `/accounts/create` - kind `create`; body type `json`; required record
  fields `ownerId`, `domain`; accepted fields `customProperties`, `domain`, `name`, `ownerId`,
  `targeted`; risk: creates a new Drift account (company) record; low-risk external mutation, no
  approval required.
- `update_account`: PATCH `/accounts/update` - kind `update`; body type `json`; required record
  fields `accountId`, `ownerId`; accepted fields `accountId`, `customProperties`, `domain`, `name`,
  `ownerId`, `targeted`; risk: mutates an existing Drift account's
  owner/name/domain/targeting/custom properties; external mutation, approval required.
- `delete_account`: DELETE `/accounts/{{ record.account_id }}` - kind `delete`; body type `none`;
  path fields `account_id`; required record fields `account_id`; accepted fields `account_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: permanently
  removes a Drift account record; destructive, approval required.
- `create_message`: POST `/conversations/{{ record.conversation_id }}/messages` - kind `create`;
  body type `json`; path fields `conversation_id`; required record fields `conversation_id`, `type`;
  accepted fields `body`, `buttons`, `conversation_id`, `type`, `userId`; confirmation
  `destructive`; risk: posts a message into a live Drift conversation, visible to the end customer
  when type is chat; external mutation, approval required.
- `create_conversation`: POST `/conversations/new` - kind `create`; body type `json`; required
  record fields `email`; accepted fields `email`; risk: starts a new Drift conversation for the
  given contact email; external mutation, approval required.
- `gdpr_retrieve`: POST `/gdpr/retrieve` - kind `custom`; body type `json`; required record fields
  `email`; accepted fields `email`; risk: triggers Drift to compile and email all data held for the
  given email address to the account's admin; a data-subject-access-request action, approval
  required.
- `gdpr_delete`: POST `/gdpr/delete` - kind `delete`; body type `json`; required record fields
  `email`; accepted fields `email`; confirmation `destructive`; risk: permanently erases every
  contact/user record matching the given email address from Drift; irreversible data-subject-erasure
  action, approval required.

## Known limits

- Batch defaults: read_page_size=65.
- API coverage includes 5 stream-backed endpoint group(s), 11 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=1, duplicate_of=7, non_data_endpoint=2, out_of_scope=5,
  requires_elevated_scope=6.
