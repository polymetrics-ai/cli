---
name: pm-intercom
description: Intercom connector knowledge and safe action guide.
---

# pm-intercom

## Purpose

Reads Intercom contacts, companies, conversations, admins, and tags through the Intercom REST API.

## Icon

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
- access_token (secret)

## ETL Streams

- contacts:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), email(), external_id(), id(), last_seen_at(), name(), owner_id(), phone(), role(), signed_up_at(), type(), unsubscribed_from_emails(), updated_at()
- companies:
  - primary key: id
  - cursor: updated_at
  - fields: company_id(), created_at(), id(), industry(), last_request_at(), monthly_spend(), name(), session_count(), size(), type(), updated_at(), user_count(), website()
- conversations:
  - primary key: id
  - cursor: updated_at
  - fields: admin_assignee_id(), created_at(), id(), open(), priority(), read(), snoozed_until(), state(), title(), type(), updated_at(), waiting_since()
- admins:
  - primary key: id
  - fields: away_mode_enabled(), away_mode_reassign(), email(), has_inbox_seat(), id(), job_title(), name(), type()
- tags:
  - primary key: id
  - fields: id(), name(), type()

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
