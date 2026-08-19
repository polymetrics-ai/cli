---
name: pm-huntr
description: Huntr connector knowledge and safe action guide.
---

# pm-huntr

## Purpose

Reads Huntr organization members, candidates, activities, notes, and actions through the Huntr REST API.

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
- page_size
- api_key (secret) (required)

## ETL Streams

- members:
  - primary key: id
  - fields: boardIds(array), createdAt(number), email(string), familyName(string), fullName(string), givenName(string), id(string), isActive(boolean), lastSeenAt(number)
- candidates:
  - primary key: id
  - fields: email(string), firstName(string), id(string), lastName(string), memberId(string)
- activities:
  - primary key: id
  - fields: activityCategory(string), completed(boolean), completedAt(number), createdAt(number), id(string), startAt(number), title(string)
- notes:
  - primary key: id
  - fields: htmlText(string), id(string), memberId(string), text(string)
- actions:
  - primary key: id
  - fields: actionType(string), activityId(string), candidateId(string), createdAt(number), date(number), id(string), memberId(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Huntr organization API read of member, candidate, activity, note, and action data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect huntr
```

### Inspect as structured JSON

```bash
pm connectors inspect huntr --json
```

## Agent Rules

- Run pm connectors inspect huntr before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
