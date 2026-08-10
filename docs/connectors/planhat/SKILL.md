---
name: pm-planhat
description: Planhat connector knowledge and safe action guide.
---

# pm-planhat

## Purpose

Reads Planhat companies, end users, and licenses through the Planhat REST API.

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
- max_pages
- mode
- page_size
- api_token (secret) (required)

## ETL Streams

- companies:
  - primary key: id
  - cursor: updated_at
  - fields: email(string), id(string), name(string), phase(string), updated_at(string)
- endusers:
  - primary key: id
  - cursor: updated_at
  - fields: email(string), id(string), name(string), phase(string), updated_at(string)
- licenses:
  - primary key: id
  - cursor: updated_at
  - fields: id(string), name(string), phase(string), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Planhat API read of customer success data
- approval: none; read-only customer success platform API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect planhat
```

### Inspect as structured JSON

```bash
pm connectors inspect planhat --json
```

## Agent Rules

- Run pm connectors inspect planhat before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
