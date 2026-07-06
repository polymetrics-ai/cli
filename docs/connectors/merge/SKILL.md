---
name: pm-merge
description: Merge connector knowledge and safe action guide.
---

# pm-merge

## Purpose

Reads Merge ATS common-model objects (candidates, applications, jobs, offers, departments, users) through the Merge unified REST API.

## Icon

- asset: icons/merge.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.merge.dev/api-reference/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- page_size
- start_date
- account_token (secret)
- api_token (secret)

## ETL Streams

- candidates:
  - primary key: id
  - cursor: modified_at
  - fields: can_email(), company(), first_name(), id(), is_private(), last_interaction_at(), last_name(), modified_at(), remote_created_at(), remote_id(), remote_updated_at(), remote_was_deleted(), title()
- applications:
  - primary key: id
  - cursor: modified_at
  - fields: applied_at(), candidate(), credited_to(), current_stage(), id(), job(), modified_at(), reject_reason(), rejected_at(), remote_id(), remote_was_deleted(), source()
- jobs:
  - primary key: id
  - cursor: modified_at
  - fields: code(), confidential(), description(), id(), modified_at(), name(), remote_created_at(), remote_id(), remote_updated_at(), remote_was_deleted(), status(), type()
- offers:
  - primary key: id
  - cursor: modified_at
  - fields: application(), closed_at(), creator(), id(), modified_at(), remote_created_at(), remote_id(), remote_was_deleted(), sent_at(), start_date(), status()
- departments:
  - primary key: id
  - cursor: modified_at
  - fields: id(), modified_at(), name(), remote_id(), remote_was_deleted()
- users:
  - primary key: id
  - cursor: modified_at
  - fields: access_role(), disabled(), email(), first_name(), id(), last_name(), modified_at(), remote_created_at(), remote_id(), remote_was_deleted()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Merge unified API read of ATS candidate and hiring data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect merge
```

### Inspect as structured JSON

```bash
pm connectors inspect merge --json
```

## Agent Rules

- Run pm connectors inspect merge before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
