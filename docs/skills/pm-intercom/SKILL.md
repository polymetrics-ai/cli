---
name: pm-intercom
description: Intercom connector knowledge and safe action guide.
---

# pm-intercom

## Purpose

Reads Intercom contacts, companies, conversations, admins, and tags through the Intercom REST API.

## Icon

- id: intercom
- asset: icons/intercom.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.intercom.com/docs/build-an-integration/learn-more/rest-apis/unversioned-changes#unversioned-changes

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- api_version
- base_url
- page_size
- access_token (secret) (required)

## ETL Streams

- contacts:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(integer), email(string), external_id(string), id(string), last_seen_at(integer), name(string), owner_id(integer), phone(string), role(string), signed_up_at(integer), type(string), unsubscribed_from_emails(boolean), updated_at(integer)
- companies:
  - primary key: id
  - cursor: updated_at
  - fields: company_id(string), created_at(integer), id(string), industry(string), last_request_at(integer), monthly_spend(number), name(string), session_count(integer), size(integer), type(string), updated_at(integer), user_count(integer), website(string)
- conversations:
  - primary key: id
  - cursor: updated_at
  - fields: admin_assignee_id(integer), created_at(integer), id(string), open(boolean), priority(string), read(boolean), snoozed_until(integer), state(string), title(string), type(string), updated_at(integer), waiting_since(integer)
- admins:
  - primary key: id
  - fields: away_mode_enabled(boolean), away_mode_reassign(boolean), email(string), has_inbox_seat(boolean), id(string), job_title(string), name(string), type(string)
- tags:
  - primary key: id
  - fields: id(string), name(string), type(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Intercom API read of contact and conversation data
- approval: none; read-only source
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect intercom
```

### Inspect as structured JSON

```bash
pm connectors inspect intercom --json
```

## Agent Rules

- Run pm connectors inspect intercom before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
