---
name: pm-lever-hiring
description: Lever Hiring connector knowledge and safe action guide.
---

# pm-lever-hiring

## Purpose

Reads Lever Hiring opportunities, postings, users, requisitions, and stages through the Lever Data API. Read-only (full-refresh).

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
- access_token (secret)
- api_key (secret)

## ETL Streams

- opportunities:
  - primary key: id
  - cursor: createdAt
  - fields: archivedAt(integer), createdAt(integer), emails(array), headline(string), id(string), lastInteractionAt(integer), name(string), origin(string), sources(array), stage(string), tags(array), updatedAt(integer)
- postings:
  - primary key: id
  - cursor: createdAt
  - fields: categories(object), createdAt(integer), hiringManager(string), id(string), owner(string), state(string), text(string), updatedAt(integer), user(string)
- users:
  - primary key: id
  - cursor: createdAt
  - fields: accessRole(string), createdAt(integer), deactivatedAt(integer), email(string), id(string), name(string), username(string)
- requisitions:
  - primary key: id
  - cursor: createdAt
  - fields: createdAt(integer), headcountHired(integer), headcountTotal(integer), id(string), name(string), owner(string), requisitionCode(string), status(string), updatedAt(integer)
- stages:
  - primary key: id
  - fields: id(string), text(string)

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
