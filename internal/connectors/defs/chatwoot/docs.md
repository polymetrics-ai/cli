# Overview

Chatwoot is an open-source (and hosted-SaaS) customer support/live-chat platform. This is a
greenfield Tier-1 declarative bundle — there is no legacy `internal/connectors/chatwoot` package to
migrate from; every stream, write action, and fixture shape here is derived directly from
Chatwoot's published OpenAPI spec (`https://developers.chatwoot.com/api-reference`, backed by
`swagger/swagger.json` in the `chatwoot/chatwoot` repo). It reads conversations, contacts, inboxes,
agents, teams, labels, and conversation-scoped messages, and writes contact/conversation/message/
label mutations through the Chatwoot Application API.

## Auth setup

Every request is scoped to one Chatwoot account via `base_url` + `account_id`
(`{{ config.base_url }}/api/v1/accounts/{{ config.account_id }}/...`). Authentication is a single
`api_access_token` request header (`mode: api_key_header`) carrying either a user's Profile Settings
personal access token or an Agent Bot token — both use the identical header name and shape per
Chatwoot's `userApiKey`/`agentBotApiKey` security schemes; this bundle does not distinguish between
them since the wire shape is identical. `account_id` is a plain (non-secret) config value: it is
visible in every dashboard URL and is not itself a credential.

## Streams notes

- **`conversations`**: `GET /conversations?status=all`, `page_number` pagination (`page` query
  param, no size param sent — Chatwoot's page size is server-fixed), records at `data.payload`,
  incremental cursor `updated_at`. `status=all` is sent explicitly because the API's own default is
  `status=open` only; a sync intended to capture the full conversation set must override it.
- **`contacts`**: `GET /contacts`, `page_number` pagination (page size documented as exactly 15 by
  Chatwoot — "Listing all the resolved contacts with pagination (Page size = 15)"), records at
  `payload`, incremental cursor `last_activity_at`. Only "resolved" contacts (identifier, email, or
  phone number present) are ever returned by this endpoint — an API-level filter, not a bundle
  simplification.
- **`inboxes`**, **`agents`**, **`teams`**: unpaginated flat listings (`inboxes`/`labels` wrap
  `payload`; `agents`/`teams` return a bare JSON array at the response root, hence `records.path:
  "."`). No incremental cursor is declared for these three streams or `labels` — Chatwoot's list
  responses for them carry no `updated_at`-shaped timestamp field to cursor on.
- **`labels`**: `GET /labels`, records at `payload`, unpaginated (account-wide label list).
- **`messages`**: conversation-scoped (`GET /conversations/{conversation_id}/messages`), so it is
  declared with a `fan_out` block: the id-listing sub-request walks `conversations` (same
  `page_number` pagination, since `ids_from.request` reuses this stream's own effective pagination
  spec per the engine dialect) to discover every conversation id, then repeats the messages
  request/pagination sequence once per id, stamping the fan-out id onto every record's
  `conversation_id` field (see `schemas/messages.json`'s note — the stamp always wins over the raw
  integer field, so `conversation_id` is declared as a string in the schema). Chatwoot's messages
  endpoint actually paginates via `before`/`after` message-id cursors (documented as returning up to
  100 messages per call), not a `page` query param at all; this bundle sends a harmless, ignored
  `page` param (inherited from the fan_out's shared pagination block) and relies on the short-page
  stop (`page_size: 25`) to end each per-conversation sub-sequence after one request in the common
  case. A conversation with more than 25 messages on its first page will still stop at page 1
  today — full `before`/`after` cursor-walking is not expressible in the current pagination dialect
  (no dialect variant reads a "next" value off the LAST message's own `id` field into a query param
  named differently from the cursor field itself); see Known limits.

## Write actions & risks

- **`create_contact`** (`POST /contacts`, JSON body) — creates a contact under a given `inbox_id`.
  Low risk; no customer notification.
- **`update_contact`** (`PUT /contacts/{id}`, JSON body, `id` carried in the path) — updates an
  existing contact's profile fields. Low risk; no customer notification.
- **`create_conversation`** (`POST /conversations`, JSON body) — creates a new conversation from a
  `source_id` (obtained via the contactable-inboxes API or a webhook event, per Chatwoot's own
  docs); customer-visible once any initial message is delivered through a live channel.
- **`send_message`** (`POST /conversations/{conversation_id}/messages`, JSON body,
  `conversation_id` carried in the path) — sends a message into an existing conversation.
  Customer-visible unless `private: true`. Only the `application/json` request shape (text
  messages) is implemented; the API's alternate `multipart/form-data` shape (file-attachment
  uploads) is out of scope — see Known limits.
- **`toggle_conversation_status`** (`POST /conversations/{conversation_id}/toggle_status`, JSON
  body, `conversation_id` carried in the path) — sets a conversation's status
  (open/resolved/pending/snoozed). Affects agent routing and reporting metrics.
- **`create_label`** (`POST /labels`, JSON body) — creates an account-wide label. Low risk; visible
  to all agents in the sidebar when `show_on_sidebar` is true.

All six actions are external mutations requiring reverse-ETL plan approval before execution.

## Known limits

- This is a curated core-surface build, not full Chatwoot API-surface parity: canned responses,
  automation rules, custom attribute definitions, custom filters, webhooks, integrations, help-center
  portals, agent bots, team-membership management, the Platform API (`account_users`), and the v2
  reporting/analytics endpoints are all out of scope. See `api_surface.json`'s `excluded` entries for
  the full per-endpoint accounting and category for each.
- `messages` fan-out pagination sends an inert `page` query parameter to an endpoint whose real
  pagination mechanism is `before`/`after` message-id cursors; a conversation whose first page of
  messages is exactly `page_size` (25) or larger will not have its remaining messages fetched by this
  bundle. Full `before`/`after` support is an `ENGINE_GAP` — the pagination dialect has no variant
  that reads the next cursor from the LAST record's own id field into a differently-named query
  param without also requiring `has_more`/`stop_path` semantics Chatwoot's messages endpoint does not
  expose.
- `create_conversation` requires the caller to already have a valid `source_id` for the target
  inbox (obtained out-of-band via the contactable-inboxes API or a webhook payload, per Chatwoot's
  own documentation) — this bundle does not orchestrate that lookup.
- `send_message`'s multipart file-attachment upload path (`attachments[]` form fields) is not
  implemented; only plain-text/JSON message bodies are supported by `send_message`.
- No stream in this bundle declares an incremental cursor for `inboxes`, `agents`, `teams`, or
  `labels` — Chatwoot's list responses for these resource types carry no timestamp field to cursor
  on, so every sync of these four streams is necessarily a full refresh.
