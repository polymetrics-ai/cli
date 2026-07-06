---
name: pm-lever-hiring
description: Lever Hiring connector knowledge and safe action guide.
---

# pm-lever-hiring

## Purpose

Reads Lever Hiring opportunities, postings, users, requisitions, and stages through the Lever Data API. Read-only (full-refresh).

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
- access_token (secret)
- api_key (secret)

## ETL Streams

- opportunities:
  - primary key: id
  - cursor: createdAt
  - fields: archivedAt(), createdAt(), emails(), headline(), id(), lastInteractionAt(), name(), origin(), sources(), stage(), tags(), updatedAt()
- postings:
  - primary key: id
  - cursor: createdAt
  - fields: categories(), createdAt(), hiringManager(), id(), owner(), state(), text(), updatedAt(), user()
- users:
  - primary key: id
  - cursor: createdAt
  - fields: accessRole(), createdAt(), deactivatedAt(), email(), id(), name(), username()
- requisitions:
  - primary key: id
  - cursor: createdAt
  - fields: createdAt(), headcountHired(), headcountTotal(), id(), name(), owner(), requisitionCode(), status(), updatedAt()
- stages:
  - primary key: id
  - fields: id(), text()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Lever API read of candidate and hiring pipeline data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect lever-hiring
```

### Inspect as structured JSON

```bash
pm connectors inspect lever-hiring --json
```

## Agent Rules

- Run pm connectors inspect lever-hiring before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
