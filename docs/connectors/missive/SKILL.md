---
name: pm-missive
description: Missive connector knowledge and safe action guide.
---

# pm-missive

## Purpose

Reads Missive contacts, contact groups, users, teams, and shared labels through the Missive REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- kind
- api_key (secret) (required)

## ETL Streams

- contacts:
  - primary key: id
  - fields: first_name(string), id(string), last_name(string), modified_at(integer)
- contact_groups:
  - primary key: id
  - fields: id(string), kind(string), name(string)
- users:
  - primary key: id
  - fields: email(string), id(string), name(string)
- teams:
  - primary key: id
  - fields: id(string), name(string), organization(string)
- shared_labels:
  - primary key: id
  - fields: color(string), id(string), name(string), name_with_parent_names(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

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
