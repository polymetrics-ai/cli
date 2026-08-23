---
name: pm-trello
description: Trello connector knowledge and safe action guide.
---

# pm-trello

## Purpose

Reads Trello boards, lists, and checklists through the Trello REST API. Cards and actions are blocked (see docs.md Known limits).

## Icon

- id: trello
- asset: icons/trello.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.atlassian.com/cloud/trello/rest/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- board_ids
- key (secret) (required)
- token (secret) (required)

## ETL Streams

- boards:
  - primary key: id
  - fields: closed(boolean), dateLastActivity(string), desc(string), id(string), idOrganization(string), name(string), shortUrl(string), url(string)
- lists:
  - primary key: id
  - fields: closed(boolean), id(string), idBoard(string), name(string), pos(number), subscribed(boolean)
- checklists:
  - primary key: id
  - fields: id(string), idBoard(string), idCard(string), name(string), pos(number)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Trello API read of board/list/checklist data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect trello
```

### Inspect as structured JSON

```bash
pm connectors inspect trello --json
```

## Agent Rules

- Run pm connectors inspect trello before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
