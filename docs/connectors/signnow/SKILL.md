---
name: pm-signnow
description: signNow connector knowledge and safe action guide.
---

# pm-signnow

## Purpose

Reads signNow documents, templates, and users through the signNow REST API.

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
- page_size
- access_token (secret)

## ETL Streams

- documents:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), updated_at()
- templates:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), updated_at()
- users:
  - primary key: id
  - cursor: updated_at
  - fields: email(), id(), name(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external signNow API read of document, template, and user data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect signnow
```

### Inspect as structured JSON

```bash
pm connectors inspect signnow --json
```

## Agent Rules

- Run pm connectors inspect signnow before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
