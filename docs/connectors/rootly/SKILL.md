---
name: pm-rootly
description: Rootly connector knowledge and safe action guide.
---

# pm-rootly

## Purpose

Reads Rootly incidents, services, and users through the Rootly API. Read-only. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

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
- start_date
- api_key (secret)

## ETL Streams

- incidents:
  - primary key: id
  - fields: id(), status(), title()
- services:
  - primary key: id
  - fields: id(), status(), title()
- users:
  - primary key: id
  - fields: id(), status(), title()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Rootly API reads performed by the legacy connector via a Tier-2 hook
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
