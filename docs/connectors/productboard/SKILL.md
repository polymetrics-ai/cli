---
name: pm-productboard
description: Productboard connector knowledge and safe action guide.
---

# pm-productboard

## Purpose

Reads Productboard features, notes, components, and products through the public API.

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
- start_date
- access_token (secret)

## ETL Streams

- features:
  - primary key: id
  - fields: created_at(), id(), name(), status(), title(), updated_at()
- notes:
  - primary key: id
  - fields: created_at(), id(), name(), status(), title(), updated_at()
- components:
  - primary key: id
  - fields: created_at(), id(), name(), status(), title(), updated_at()
- products:
  - primary key: id
  - fields: created_at(), id(), name(), status(), title(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Productboard API read of feature, note, component, and product data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect productboard
```

### Inspect as structured JSON

```bash
pm connectors inspect productboard --json
```

## Agent Rules

- Run pm connectors inspect productboard before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
