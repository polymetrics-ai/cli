---
name: pm-vitally
description: Vitally connector knowledge and safe action guide.
---

# pm-vitally

## Purpose

Reads and writes Vitally customer-success accounts, users, notes, conversations, tasks, and NPS responses via the Vitally REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- page_size
- status
- basic_auth_header (secret) (required)

## ETL Streams

- accounts:
  - primary key: id
  - fields: id(string), name(string), traits(object)
- users:
  - primary key: id
  - cursor: updatedAt
  - fields: accounts(array), avatar(string), createdAt(string), deactivatedAt(string), email(string), externalId(string), firstKnown(string), id(string), lastInboundMessageTimestamp(string), lastOutboundMessageTimestamp(string), lastSeenTimestamp(string), name(string), npsLastFeedback(string), npsLastRespondedAt(string), npsLastScore(integer), organizations(array), segments(array), traits(object), unsubscribedFromConversations(boolean), unsubscribedFromConversationsAt(string), updatedAt(string)
- notes:
  - primary key: id
  - cursor: updated_at
  - fields: account_id(string), archived_at(string), author_id(string), category_id(string), created_at(string), external_id(string), id(string), note(string), note_date(string), organization_id(string), source(string), subject(string), tags(array), traits(object), updated_at(string)
- conversations:
  - primary key: id
  - cursor: updated_at
  - fields: accounts(array), admins(array), created_at(string), external_id(string), id(string), rating(string), source(string), status(string), subject(string), traits(object), updated_at(string), users(array)
- tasks:
  - primary key: id
  - cursor: updated_at
  - fields: account_id(string), archived_at(string), assigned_to_id(string), category_id(string), completed_at(string), completed_by_id(string), created_at(string), description(string), due_date(string), external_id(string), id(string), meeting_id(string), name(string), organization_id(string), source(string), tags(array), traits(object), updated_at(string)
- nps_responses:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), external_id(string), feedback(string), id(string), responded_at(string), score(integer), updated_at(string), user_id(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_account:
  - endpoint: POST /resources/accounts
  - required fields: externalId, name
  - risk: creates a new customer-success account visible to the vendor's CS team; external mutation, approval required
- update_account:
  - endpoint: PUT /resources/accounts/{{ record.id }}
  - required fields: id
  - risk: updates an existing customer-success account's fields/traits, visible to the vendor's CS team; external mutation, approval required
- create_user:
  - endpoint: POST /resources/users
  - required fields: externalId
  - risk: creates a new user record visible to the vendor's CS team; external mutation, approval required
- update_user:
  - endpoint: PUT /resources/users/{{ record.id }}
  - required fields: id
  - risk: updates an existing user's fields/traits, visible to the vendor's CS team; external mutation, approval required
- create_note:
  - endpoint: POST /resources/notes
  - required fields: note, noteDate
  - risk: creates a customer-success note visible to the vendor's CS team; external mutation, approval required
- update_note:
  - endpoint: PUT /resources/notes/{{ record.id }}
  - required fields: id
  - risk: updates an existing customer-success note visible to the vendor's CS team; external mutation, approval required
- delete_note:
  - endpoint: DELETE /resources/notes/{{ record.id }}
  - required fields: id
  - risk: archives/deletes a customer-success note; external mutation, approval required
- create_conversation:
  - endpoint: POST /resources/conversations
  - required fields: subject, messages
  - risk: creates a historical conversation record visible to the vendor's CS team; does not send outbound messages to real participants (Vitally's own documented behavior); external mutation, approval required
- update_conversation:
  - endpoint: PUT /resources/conversations/{{ record.id }}
  - required fields: id
  - risk: updates an existing conversation record (new messages inserted, existing ones updated by externalId); external mutation, approval required
- delete_conversation:
  - endpoint: DELETE /resources/conversations/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a conversation and all its messages; external mutation, approval required
- create_task:
  - endpoint: POST /resources/tasks
  - required fields: name, accountId
  - risk: creates a customer-success task visible to the vendor's CS team; external mutation, approval required
- update_task:
  - endpoint: PUT /resources/tasks/{{ record.id }}
  - required fields: id
  - risk: updates an existing customer-success task visible to the vendor's CS team; external mutation, approval required
- create_nps_response:
  - endpoint: POST /resources/npsResponses
  - required fields: userId, respondedAt, score
  - risk: creates (or, if externalId already exists, upserts -- Vitally's own documented behavior) an NPS response visible to the vendor's CS team; external mutation, approval required
- update_nps_response:
  - endpoint: PUT /resources/npsResponses/{{ record.id }}
  - required fields: id
  - risk: updates an existing NPS response visible to the vendor's CS team; external mutation, approval required

## Security

- read risk: external Vitally API read of customer-success account/user/note/conversation/task/NPS-response data
- write risk: external mutation of Vitally customer-success records (create/update accounts, users, notes, tasks, conversations, NPS responses; delete notes and conversations); approval required
- approval: read: none, read-only sync surface. write: required for all mutating actions (create/update/delete of customer-success records visible to the vendor's CS team).
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect vitally
```

### Inspect as structured JSON

```bash
pm connectors inspect vitally --json
```

## Agent Rules

- Run pm connectors inspect vitally before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
