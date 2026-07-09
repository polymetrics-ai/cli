# Overview

Reads Chatwoot support conversations, contacts, inboxes, agents, teams, labels, and
conversation-scoped messages, and writes approved contact/conversation/message/label mutations
through the account-scoped Chatwoot Application API. The connector also tracks the full official
Swagger inventory as blocked-by-default operation metadata so later direct-read, binary/file, and
sensitive/admin reverse-ETL slices can add typed commands without exposing a raw generic API tool.

Readable streams: `conversations`, `contacts`, `inboxes`, `agents`, `teams`, `labels`, `messages`.

Write actions: `create_contact`, `update_contact`, `create_conversation`, `send_message`,
`toggle_conversation_status`, `create_label`.

Service API documentation: https://developers.chatwoot.com/api-reference. Official inventory source:
https://raw.githubusercontent.com/chatwoot/chatwoot/develop/swagger/swagger.json.

## Auth setup

Connection fields:

- `account_id` (required, string); The numeric Chatwoot account ID to scope every request to
  (visible in the dashboard URL, e.g. app.chatwoot.com/app/accounts/1/...).
- `api_access_token` (optional, secret, string); Chatwoot API access token (Profile Settings >
  Access Token for a user token, or an Agent Bot token). Sent as the api_access_token request header
  on every request.
- `base_url` (required, string); format `uri`; Your Chatwoot instance root, e.g.
  https://app.chatwoot.com for the hosted SaaS or your self-hosted install's origin. The engine
  appends /api/v1/accounts/{account_id} (and, for messages, the conversation-scoped sub-path) to
  every request; do not include /api/v1 yourself. Also usable as a base URL override for
  tests/proxies.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound for the first incremental
  sync of conversations (e.g. 2026-01-01T00:00:00Z). Contacts use their persisted cursor when one
  exists. Ignored once a cursor has been persisted.

Secret fields are redacted in logs and write previews: `api_access_token`.

Authentication behavior:

- API key authentication in `api_access_token` using `secrets.api_access_token`.

Requests use base URL `{{ config.base_url }}/api/v1/accounts/{{ config.account_id }}` after applying
configuration defaults.

Connection checks call GET `/agents`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `inboxes`, `agents`, `teams`, `labels`; page_number: `conversations`,
`contacts`, `messages`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `conversations`: GET `/conversations` - records path `data.payload`; query `status`=`all`;
  page-number pagination; page parameter `page`; no page-size parameter; starts at 1; page size 25;
  incremental cursor `updated_at`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `contacts`: GET `/contacts` - records path `payload`; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 15; incremental cursor `last_activity_at`;
  formatted as `rfc3339`.
- `inboxes`: GET `/inboxes` - records path `payload`.
- `agents`: GET `/agents` - records path `.`.
- `teams`: GET `/teams` - records path `.`.
- `labels`: GET `/labels` - records path `payload`.
- `messages`: GET `/conversations/{{ fanout.id }}/messages` - records path `payload`; page-number
  pagination; page parameter `page`; no page-size parameter; starts at 1; page size 25; fan-out; ids
  from request `/conversations`; id-list records path `data.payload`; id field `id`; id inserted
  into the request path; stamps `conversation_id`.

## Write actions & risks

Overall write risk: external mutation of Chatwoot contacts, conversations, messages, and labels;
agent-visible and customer-visible side effects.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_contact`: POST `/contacts` - kind `create`; body type `json`; required record fields
  `inbox_id`; accepted fields `additional_attributes`, `avatar_url`, `blocked`, `custom_attributes`,
  `email`, `identifier`, `inbox_id`, `name`, `phone_number`; risk: creates a new Chatwoot contact
  record; low risk, no customer notification.
- `update_contact`: PUT `/contacts/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `additional_attributes`, `avatar_url`,
  `blocked`, `custom_attributes`, `email`, `id`, `identifier`, `name`, `phone_number`; risk: updates
  an existing Chatwoot contact's profile fields; low risk, no customer notification.
- `create_conversation`: POST `/conversations` - kind `create`; body type `json`; required record
  fields `source_id`; accepted fields `additional_attributes`, `assignee_id`, `contact_id`,
  `custom_attributes`, `inbox_id`, `source_id`, `status`, `team_id`; risk: creates a new
  conversation in the target inbox; customer-visible once the initial message is delivered through a
  live channel.
- `send_message`: POST `/conversations/{{ record.conversation_id }}/messages` - kind `create`; body
  type `json`; path fields `conversation_id`; required record fields `conversation_id`, `content`;
  accepted fields `content`, `content_attributes`, `content_type`, `conversation_id`,
  `message_type`, `private`; risk: sends a message into a conversation; customer-visible unless
  private is true and may notify the contact through the inbox channel.
- `toggle_conversation_status`: POST `/conversations/{{ record.conversation_id }}/toggle_status` -
  kind `update`; body type `json`; path fields `conversation_id`; required record fields
  `conversation_id`, `status`; accepted fields `conversation_id`, `snoozed_until`, `status`; risk:
  changes a conversation's status (open/resolved/pending/snoozed); may affect agent routing and
  reporting metrics.
- `create_label`: POST `/labels` - kind `create`; body type `json`; required record fields `title`;
  accepted fields `color`, `description`, `show_on_sidebar`, `title`; risk: creates a new
  account-wide label; low risk, visible to all agents in the sidebar when show_on_sidebar is true.

## Known limits

- Batch defaults: read_page_size=15.
- The official Swagger inventory currently accounts for 144 operations across 89 paths: GET 62,
  POST 41, PATCH 21, DELETE 18, PUT 2.
- Executable coverage in this slice remains 7 stream-backed endpoints and 6 write-backed endpoints.
  The remaining official operations are blocked-by-default operation metadata: direct_read=53,
  admin_reverse_etl=35, sensitive_reverse_etl=19, destructive_action=19, disallowed=4, duplicate=1.
- Direct-read commands for single records, search/filter endpoints, reports, audit logs, and public
  inbox resources are planned but not executable until typed bounded output policies exist.
- Admin/configuration and destructive operations are not blanket-excluded. They remain blocked until
  later slices add named reverse-ETL actions with exact schemas, risk text, preview, approval, and
  typed confirmation where required.
- Message attachment and profile/avatar multipart variants are not implemented by the current JSON
  write actions; they require the later binary/file policy slice.
