---
name: pm-opuswatch
description: OPUSWatch connector knowledge and safe action guide.
---

# pm-opuswatch

## Purpose

Reads OPUSWatch monitors, incidents, and checks.

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
- mode
- page_size
- api_key (secret) (required)

## ETL Streams

- monitors:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(string), message(string), name(string), status(string), updated_at(string)
- incidents:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(string), message(string), name(string), status(string), updated_at(string)
- checks:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(string), message(string), name(string), status(string), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external OPUSWatch API read of monitor, incident, and check status data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect opuswatch
```

### Inspect as structured JSON

```bash
pm connectors inspect opuswatch --json
```

## Agent Rules

- Run pm connectors inspect opuswatch before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
