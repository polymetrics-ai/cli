---
name: pm-canny
description: Canny connector knowledge and safe action guide.
---

# pm-canny

## Purpose

Reads Canny boards, posts, comments, categories, and companies through fixed Canny REST form requests.

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

- api_key (secret) (required)

## ETL Streams

- boards:
  - primary key: id
  - cursor: created
  - fields: created(string), id(string)
- posts:
  - primary key: id
  - cursor: created
  - fields: created(string), id(string)
- comments:
  - primary key: id
  - cursor: created
  - fields: created(string), id(string)
- categories:
  - primary key: id
  - cursor: created
  - fields: created(string), id(string)
- companies:
  - primary key: id
  - cursor: created
  - fields: created(string), id(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: Bounded Canny list requests carry the declared API key only in typed form bodies.
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect canny
```

### Inspect as structured JSON

```bash
pm connectors inspect canny --json
```

## Agent Rules

- Run pm connectors inspect canny before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
