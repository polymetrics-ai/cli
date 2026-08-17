---
name: pm-paperform
description: Paperform connector knowledge and safe action guide.
---

# pm-paperform

## Purpose

Reads Paperform forms and form submissions through the Paperform REST API.

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
- form_id
- api_key (secret)

## ETL Streams

- forms:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), id(), slug(), title(), updated_at()
- submissions:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), data(), form_id(), id(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Paperform API read of form and submission data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect paperform
```

### Inspect as structured JSON

```bash
pm connectors inspect paperform --json
```

## Agent Rules

- Run pm connectors inspect paperform before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
