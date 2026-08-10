# Overview

Reads and writes the account-scoped Chatwoot Application API support-desk surface: ETL streams for
conversations, contacts, inboxes, agents, teams, labels, and conversation-scoped messages, plus
bounded direct reads and reverse-ETL actions for agent bots, canned responses, custom attribute
definitions, custom filters, webhooks, integration hooks, automation rules, help-center portals,
inbox membership, and account settings. `api_surface.json` owns source-backed endpoint accounting;
`cli_surface.json` owns command availability. See `pm chatwoot` (bare) and
`pm chatwoot <group> --help` for the current available command surface, or
`docs/connectors/chatwoot/MANUAL.md` for the generated reference.

Readable streams: `conversations`, `contacts`, `inboxes`, `agents`, `teams`, `labels`, `messages`.

The six write actions below were this connector's original build. `writes.json` owns the complete
action set; consult `cli_surface.json` or runtime help before assuming a namespace command is
available for a declared action.

Service API documentation: https://developers.chatwoot.com/api-reference.

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
  sync of conversations, contacts, and messages (e.g. 2026-01-01T00:00:00Z). Ignored once a cursor
  has been persisted.

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
- `api_surface.json` owns the Chatwoot operation inventory, artifact provenance, and endpoint
  disposition. `cli_surface.json` owns command availability; a `covered_by.write` endpoint means
  there is a declared write action, not necessarily an executable namespace command.
- Blocked or unsupported operations fall outside this bundle's single account-scoped base URL
  (`/api/v1/accounts/{account_id}`): Chatwoot's `/api/v2` reporting surface, `/platform/api/v1`
  (a distinct, more-privileged credential), `/public/api/v1` (an anonymous widget-actor API),
  `/survey`, and `/api/v1/profile`. One inbox-provisioning pair (`POST`/`PATCH .../inboxes`) is
  blocked because its record schema is a genuine `oneOf` across ~10 channel types. One operation
  (`GET /platform/api/v1/users/{id}/login`) is permanently unsupported because it returns a
  credential-derived SSO/impersonation URL.
- `conversations update-custom-attributes` has `availability: partial`: its required
  `custom_attributes` field is an arbitrary-keyed JSON object that flat CLI flags cannot express;
  use a reverse-ETL plan JSON record override instead.
