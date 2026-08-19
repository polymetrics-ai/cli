---
name: pm-gologin
description: GoLogin connector knowledge and safe action guide.
---

# pm-gologin

## Purpose

Reads GoLogin browser profiles, folders, tags, and account information through the GoLogin REST API.

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
- api_key (secret) (required)

## ETL Streams

- profiles:
  - primary key: id
  - cursor: updatedAt
  - fields: browserType(string), createdAt(string), folderName(string), id(string), name(string), notes(string), os(string), role(string), updatedAt(string)
- folders:
  - primary key: id
  - fields: id(string), name(string), profilesCount(integer)
- user:
  - primary key: _id
  - cursor: createdAt
  - fields: _id(string), createdAt(string), email(string), firstName(string), lastName(string), plan(string), profilesCount(integer)
- tags:
  - primary key: _id
  - fields: _id(string), color(string), field(string), title(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external GoLogin API read of browser profile and account data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect gologin
```

### Inspect as structured JSON

```bash
pm connectors inspect gologin --json
```

## Agent Rules

- Run pm connectors inspect gologin before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
