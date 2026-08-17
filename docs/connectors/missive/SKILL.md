---
name: pm-missive
description: Missive connector knowledge and safe action guide.
---

# pm-missive

## Purpose

Reads Missive contacts, contact groups, users, teams, and shared labels through the Missive REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- kind
- api_key (secret)

## ETL Streams

- contacts:
  - primary key: id
  - fields: first_name(), id(), last_name(), modified_at()
- contact_groups:
  - primary key: id
  - fields: id(), kind(), name()
- users:
  - primary key: id
  - fields: email(), id(), name()
- teams:
  - primary key: id
  - fields: id(), name(), organization()
- shared_labels:
  - primary key: id
  - fields: color(), id(), name(), name_with_parent_names()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Missive API read of contact, user, team, and label data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect missive
```

### Inspect as structured JSON

```bash
pm connectors inspect missive --json
```

## Agent Rules

- Run pm connectors inspect missive before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
