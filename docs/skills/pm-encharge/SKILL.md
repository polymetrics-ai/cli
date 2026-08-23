---
name: pm-encharge
description: Encharge connector knowledge and safe action guide.
---

# pm-encharge

## Purpose

Reads Encharge people, segments, fields, account tags, and schemas through the Encharge REST API.

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
- api_key (secret) (required)

## ETL Streams

- peoples:
  - primary key: id
  - fields: company(string), country(string), createdAt(string), email(string), firstName(string), id(string), lastName(string), name(string), phone(string), title(string), updatedAt(string), userId(string)
- segments:
  - primary key: id
  - fields: createdAt(string), id(string), name(string), type(string), updatedAt(string)
- fields:
  - primary key: name
  - fields: format(string), name(string), title(string), type(string)
- account_tags:
  - primary key: tag
  - fields: createdAt(string), id(string), tag(string)
- schemas:
  - primary key: name
  - fields: name(string), title(string), type(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Encharge API read of people, segment, field, and tag data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect encharge
```

### Inspect as structured JSON

```bash
pm connectors inspect encharge --json
```

## Agent Rules

- Run pm connectors inspect encharge before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
