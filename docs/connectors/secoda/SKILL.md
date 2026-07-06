---
name: pm-secoda
description: Secoda connector knowledge and safe action guide.
---

# pm-secoda

## Purpose

Reads Secoda catalog metadata (tables, documents, collections, questions) through the Secoda API.

## Icon

- asset: icons/secoda.svg
- source: official
- review_status: official_verified
- review_url: https://docs.secoda.co/api.md

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret)

## ETL Streams

- tables:
  - primary key: id
  - fields: id(), name(), updated_at()
- documents:
  - primary key: id
  - fields: id(), name(), updated_at()
- collections:
  - primary key: id
  - fields: id(), name(), updated_at()
- questions:
  - primary key: id
  - fields: id(), name(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Secoda API read of data-catalog metadata
- approval: none; read-only, no reverse-ETL writes implemented by legacy
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect secoda
```

### Inspect as structured JSON

```bash
pm connectors inspect secoda --json
```

## Agent Rules

- Run pm connectors inspect secoda before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
