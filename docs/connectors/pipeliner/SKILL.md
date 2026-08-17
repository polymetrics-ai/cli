---
name: pm-pipeliner
description: Pipeliner connector knowledge and safe action guide.
---

# pm-pipeliner

## Purpose

Reads Pipeliner CRM accounts, contacts, opportunities, and leads through the REST API.

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
- mode
- space_id
- password (secret)
- username (secret)

## ETL Streams

- accounts:
  - primary key: id
  - fields: id(), name(), status(), updated_at()
- contacts:
  - primary key: id
  - fields: id(), name(), status(), updated_at()
- opportunities:
  - primary key: id
  - fields: id(), name(), status(), updated_at()
- leads:
  - primary key: id
  - fields: id(), name(), status(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Pipeliner CRM API read of account, contact, opportunity, and lead data
- approval: none; read-only CRM sync
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect pipeliner
```

### Inspect as structured JSON

```bash
pm connectors inspect pipeliner --json
```

## Agent Rules

- Run pm connectors inspect pipeliner before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
