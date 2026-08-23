---
name: pm-drift
description: Drift connector knowledge and safe action guide.
---

# pm-drift

## Purpose

Reads Drift users, accounts, conversations, contacts, and teams, and writes contact/account/message/conversation/timeline-event/GDPR mutations through the Drift REST API.

## Icon

- id: drift
- asset: icons/drift.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://devdocs.drift.com/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- email
- access_token (secret) (required)

## ETL Streams

- users:
  - primary key: id
  - cursor: updatedAt
  - fields: alias(string), availability(string), avatarUrl(string), bot(boolean), createdAt(integer), email(string), id(integer), locale(string), name(string), orgId(integer), phone(string), role(string), timeZone(string), updatedAt(integer), verified(boolean)
- accounts:
  - primary key: account_id
  - cursor: updateDateTime
  - fields: account_id(string), createDateTime(integer), customProperties(array), deleted(boolean), domain(string), name(string), ownerId(integer), targeted(boolean), updateDateTime(integer)
- conversations:
  - primary key: id
  - cursor: updatedAt
  - fields: contactId(integer), conversationTags(array), createdAt(integer), id(integer), inboxId(integer), orgId(integer), participants(array), relatedPlaybookId(integer), status(string), updatedAt(integer)
- contacts:
  - primary key: id
  - cursor: updatedAt
  - fields: attributes(object), createdAt(integer), id(integer), updatedAt(integer)
- teams:
  - primary key: id
  - cursor: updatedAt
  - fields: autoOffline(boolean), id(integer), main(boolean), members(array), name(string), orgId(integer), owner(integer), responseTimerEnabled(boolean), status(string), teamAvailabilityMode(string), updatedAt(integer), workspaceId(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_contact:
  - endpoint: POST /contacts
  - required fields: attributes
  - risk: creates a new Drift contact record; low-risk external mutation, no approval required
- update_contact:
  - endpoint: PATCH /contacts/{{ record.id }}
  - required fields: id, attributes
  - risk: mutates an existing Drift contact's attributes, including standard fields (email/name/phone) and any custom attribute; external mutation, approval required
- delete_contact:
  - endpoint: DELETE /contacts/{{ record.id }}
  - required fields: id
  - risk: permanently removes a Drift contact and its conversation history association; destructive, approval required
- post_timeline_event:
  - endpoint: POST /contacts/timeline
  - required fields: contactId, event
  - risk: posts a custom timeline event onto a contact's record; low-risk external mutation, no approval required
- create_account:
  - endpoint: POST /accounts/create
  - required fields: ownerId, domain
  - risk: creates a new Drift account (company) record; low-risk external mutation, no approval required
- update_account:
  - endpoint: PATCH /accounts/update
  - required fields: accountId, ownerId
  - risk: mutates an existing Drift account's owner/name/domain/targeting/custom properties; external mutation, approval required
- delete_account:
  - endpoint: DELETE /accounts/{{ record.account_id }}
  - required fields: account_id
  - risk: permanently removes a Drift account record; destructive, approval required
- create_message:
  - endpoint: POST /conversations/{{ record.conversation_id }}/messages
  - required fields: conversation_id, type
  - risk: posts a message into a live Drift conversation, visible to the end customer when type is chat; external mutation, approval required
- create_conversation:
  - endpoint: POST /conversations/new
  - required fields: email
  - risk: starts a new Drift conversation for the given contact email; external mutation, approval required
- gdpr_retrieve:
  - endpoint: POST /gdpr/retrieve
  - required fields: email
  - risk: triggers Drift to compile and email all data held for the given email address to the account's admin; a data-subject-access-request action, approval required
- gdpr_delete:
  - endpoint: POST /gdpr/delete
  - required fields: email
  - risk: permanently erases every contact/user record matching the given email address from Drift; irreversible data-subject-erasure action, approval required

## Security

- read risk: external Drift API read of conversational-marketing users, accounts, conversations, contacts, and teams
- write risk: external Drift API mutation of contacts, accounts, conversations, messages, timeline events, and GDPR data-subject requests; delete_contact/delete_account/gdpr_delete are destructive and require approval
- approval: required for delete_contact, delete_account, and gdpr_delete; other create/update writes are lower-risk marketing/support-data mutations
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect drift
```

### Inspect as structured JSON

```bash
pm connectors inspect drift --json
```

## Agent Rules

- Run pm connectors inspect drift before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
