# Overview

Reads and writes Vitally customer-success accounts, users, notes, conversations, tasks, and NPS
responses via the Vitally REST API.

Readable streams: `accounts`, `users`, `notes`, `conversations`, `tasks`, `nps_responses`.

Write actions: `create_account`, `update_account`, `create_user`, `update_user`, `create_note`,
`update_note`, `delete_note`, `create_conversation`, `update_conversation`, `delete_conversation`,
`create_task`, `update_task`, `create_nps_response`, `update_nps_response`.

Service API documentation: https://docs.vitally.io/en/articles/9880649-rest-api-overview.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://rest.vitally.io`; format `uri`; Vitally REST API
  base URL override for tests or proxies.
- `basic_auth_header` (required, secret, string); The full, pre-built Authorization header value
  Vitally expects (e.g. 'Basic <base64 apiKey:>').
- `page_size` (optional, integer); default `100`; Vitally's documented max/default is 100.
- `status` (optional, string); allowed values `active`, `churned`, `activeOrChurned`; Vitally
  defaults to 'active' server-side when this is left unset.

Secret fields are redacted in logs and write previews: `basic_auth_header`.

Default configuration values: `base_url=https://rest.vitally.io`, `page_size=100`.

Authentication behavior:

- API key authentication in `Authorization` using `secrets.basic_auth_header`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/resources/accounts`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `from`; next token from `next`. Responses
return records in a `results` array alongside a `next` cursor; a `next` value of `null` marks the
last page.

All streams are full-refresh only: Vitally's list endpoints support `sortBy` ordering but no
since/after filters, so incremental sync is unavailable.

Pagination by stream: cursor: `users`, `notes`, `conversations`, `tasks`, `nps_responses`; none:
`accounts`.

- `accounts`: GET `/resources/accounts` - records path `results`; query `status` from template `{{
  config.status }}`, omitted when absent.
- `users`: GET `/resources/users` - records path `results`; query `limit` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `from`; next token from
  `next`.
- `notes`: GET `/resources/notes` - records path `results`; query `limit` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `from`; next token from
  `next`; computed output fields `account_id`, `archived_at`, `author_id`, `category_id`,
  `created_at`, `external_id`, `note_date`, `organization_id`, `updated_at`.
- `conversations`: GET `/resources/conversations` - records path `results`; query `limit` from
  template `{{ config.page_size }}`, default `100`; cursor pagination; cursor parameter `from`; next
  token from `next`; computed output fields `created_at`, `external_id`, `updated_at`.
- `tasks`: GET `/resources/tasks` - records path `results`; query `limit` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `from`; next token from
  `next`; computed output fields `account_id`, `archived_at`, `assigned_to_id`, `category_id`,
  `completed_at`, `completed_by_id`, `created_at`, `due_date`, `external_id`, `organization_id`,
  `updated_at`.
- `nps_responses`: GET `/resources/npsResponses` - records path `results`; query `limit` from
  template `{{ config.page_size }}`, default `100`; cursor pagination; cursor parameter `from`; next
  token from `next`; computed output fields `created_at`, `external_id`, `responded_at`,
  `updated_at`, `user_id`.

## Write actions & risks

Overall write risk: external mutation of Vitally customer-success records (create/update accounts,
users, notes, tasks, conversations, NPS responses; delete notes and conversations); approval
required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_account`: POST `/resources/accounts` - kind `create`; body type `json`; required record
  fields `externalId`, `name`; accepted fields `externalId`, `name`, `organizationId`, `traits`;
  risk: creates a new customer-success account visible to the vendor's CS team; external mutation,
  approval required.
- `update_account`: PUT `/resources/accounts/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `accountOwnerId`, `id`, `name`,
  `organizationId`, `traits`; risk: updates an existing customer-success account's fields/traits,
  visible to the vendor's CS team; external mutation, approval required.
- `create_user`: POST `/resources/users` - kind `create`; body type `json`; required record fields
  `externalId`; accepted fields `accountIds`, `avatar`, `email`, `externalId`, `joinDate`, `name`,
  `organizationIds`, `traits`, `unsubscribedFromConversations`; risk: creates a new user record
  visible to the vendor's CS team; external mutation, approval required.
- `update_user`: PUT `/resources/users/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `avatar`, `email`, `id`, `name`,
  `traits`, `unsubscribedFromConversations`; risk: updates an existing user's fields/traits, visible
  to the vendor's CS team; external mutation, approval required.
- `create_note`: POST `/resources/notes` - kind `create`; body type `json`; required record fields
  `note`, `noteDate`; accepted fields `authorId`, `categoryId`, `externalId`, `note`, `noteDate`,
  `subject`, `tags`, `traits`; risk: creates a customer-success note visible to the vendor's CS
  team; external mutation, approval required.
- `update_note`: PUT `/resources/notes/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `authorId`, `categoryId`, `id`, `note`,
  `noteDate`, `subject`, `tags`, `traits`; note that `tags` replaces the entire tag set -- a
  partial `tags` array overwrites existing tags rather than merging; risk: updates an existing
  customer-success note visible to the vendor's CS team; external mutation, approval required.
- `delete_note`: DELETE `/resources/notes/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: archives/deletes a customer-success note;
  external mutation, approval required.
- `create_conversation`: POST `/resources/conversations` - kind `create`; body type `json`; required
  record fields `subject`, `messages`; accepted fields `externalId`, `messages`, `subject`,
  `traits`; risk: creates a historical conversation record visible to the vendor's CS team; does not
  send outbound messages to real participants (Vitally's own documented behavior); external
  mutation, approval required.
- `update_conversation`: PUT `/resources/conversations/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `id`, `messages`,
  `subject`, `traits`; risk: updates an existing conversation record (new messages inserted,
  existing ones updated by externalId); external mutation, approval required.
- `delete_conversation`: DELETE `/resources/conversations/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: permanently deletes a
  conversation and all its messages; external mutation, approval required.
- `create_task`: POST `/resources/tasks` - kind `create`; body type `json`; required record fields
  `name`, `accountId`; accepted fields `accountId`, `assignedToId`, `categoryId`, `completedAt`,
  `completedById`, `description`, `dueDate`, `externalId`, `name`, `organizationId`, `tags`,
  `traits`; risk: creates a customer-success task visible to the vendor's CS team; external
  mutation, approval required.
- `update_task`: PUT `/resources/tasks/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `accountId`, `assignedToId`,
  `categoryId`, `completedAt`, `completedById`, `description`, `dueDate`, `id`, `name`,
  `organizationId`, `tags`, `traits`; risk: updates an existing customer-success task visible to the
  vendor's CS team; external mutation, approval required.
- `create_nps_response`: POST `/resources/npsResponses` - kind `create`; body type `json`; required
  record fields `userId`, `respondedAt`, `score`; accepted fields `externalId`, `feedback`,
  `respondedAt`, `score`, `userId`; risk: creates (or, if externalId already exists, upserts --
  Vitally's own documented behavior) an NPS response visible to the vendor's CS team; external
  mutation, approval required.
- `update_nps_response`: PUT `/resources/npsResponses/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `feedback`, `id`,
  `respondedAt`, `score`; risk: updates an existing NPS response visible to the vendor's CS team;
  external mutation, approval required.

## Known limits

- API coverage includes 6 stream-backed endpoint group(s), 14 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=16, out_of_scope=3.
