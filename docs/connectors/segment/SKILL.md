---
name: pm-segment
description: Segment connector knowledge and safe action guide.
---

# pm-segment

## Purpose

Reads Segment workspace, source, and destination metadata through the Segment Public API.

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
- api_token (secret) (required)

## ETL Streams

- workspaces:
  - primary key: id
  - fields: id(string), name(string), slug(string), updated_at(string)
- sources:
  - primary key: id
  - fields: id(string), name(string), slug(string), updated_at(string)
- destinations:
  - primary key: id
  - fields: id(string), name(string), slug(string), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Segment Public API read of workspace, source, and destination metadata
- approval: none; read-only, no reverse-ETL writes implemented by legacy
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect segment
```

### Inspect as structured JSON

```bash
pm connectors inspect segment --json
```

## Agent Rules

- Run pm connectors inspect segment before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
