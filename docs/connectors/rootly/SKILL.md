---
name: pm-rootly
description: Rootly connector knowledge and safe action guide.
---

# pm-rootly

## Purpose

Reads Rootly incidents, services, and users through fixed JSON:API routes.

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

- start_date (required)
- api_key (secret) (required)

## ETL Streams

- incidents:
  - primary key: id
  - fields: id(string), status(string), title(string)
- services:
  - primary key: id
  - fields: id(string), status(string), title(string)
- users:
  - primary key: id
  - fields: id(string), status(string), title(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: Bounded Rootly JSON:API reads use the fixed provider origin and declared bearer authentication.
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect rootly
```

### Inspect as structured JSON

```bash
pm connectors inspect rootly --json
```

## Agent Rules

- Run pm connectors inspect rootly before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
