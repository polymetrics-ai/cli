---
name: pm-safetyculture
description: SafetyCulture connector knowledge and safe action guide.
---

# pm-safetyculture

## Purpose

Reads SafetyCulture audits, templates, and users through fixed REST routes.

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
- access_token (secret) (required)

## ETL Streams

- audits:
  - primary key: id
  - fields: id(string), modified_at(string), name(string)
- templates:
  - primary key: id
  - fields: id(string), modified_at(string), name(string)
- users:
  - primary key: id
  - fields: id(string), modified_at(string), name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: Bounded GET reads use the fixed SafetyCulture origin and declared bearer authentication.
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect safetyculture
```

### Inspect as structured JSON

```bash
pm connectors inspect safetyculture --json
```

## Agent Rules

- Run pm connectors inspect safetyculture before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
