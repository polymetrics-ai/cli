---
name: pm-freshcaller
description: Freshcaller connector knowledge and safe action guide.
---

# pm-freshcaller

## Purpose

Reads Freshcaller calls, agents, teams, and phone numbers through the Freshcaller REST API.

## Icon

- asset: icons/freshcaller.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.freshcaller.com/api/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- mode
- page_size
- api_key (secret)

## ETL Streams

- calls:
  - primary key: id
  - cursor: call_time
  - fields: agent_id(), call_time(), direction(), duration(), id(), phone_number(), status()
- agents:
  - primary key: id
  - fields: email(), id(), name(), status()
- teams:
  - primary key: id
  - fields: id(), name()
- numbers:
  - primary key: id
  - fields: id(), name(), phone_number()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Freshcaller API read of call, agent, team, and phone number data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect freshcaller
```

### Inspect as structured JSON

```bash
pm connectors inspect freshcaller --json
```

## Agent Rules

- Run pm connectors inspect freshcaller before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
