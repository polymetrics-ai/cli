---
name: pm-scryfall
description: Scryfall connector knowledge and safe action guide.
---

# pm-scryfall

## Purpose

Reads cards and sets from the public Scryfall API. Read-only and credential-free.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- base_url
- q

## ETL Streams

- cards:
  - primary key: id
  - fields: id(), name(), set()
- sets:
  - primary key: id
  - fields: id(), name(), set()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: public, credential-free Scryfall API read of card and set data
- approval: none; read-only, no reverse-ETL writes implemented by legacy
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect scryfall
```

### Inspect as structured JSON

```bash
pm connectors inspect scryfall --json
```

## Agent Rules

- Run pm connectors inspect scryfall before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
