---
name: pm-tinyemail
description: TinyEmail connector knowledge and safe action guide.
---

# pm-tinyemail

## Purpose

Reads subscribers, lists, and campaigns, and writes subscriber create/upsert actions, through the tinyEmail API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret)

## ETL Streams

- subscribers:
  - primary key: id
  - fields: created_at(), email(), first_name(), id(), last_name(), status()
- lists:
  - primary key: id
  - fields: created_at(), id(), name(), subscriber_count()
- campaigns:
  - primary key: id
  - fields: created_at(), id(), name(), status(), subject()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_subscriber:
  - endpoint: POST /segment/customer
  - risk: creates or upserts a subscriber (customer) into the caller's tinyEmail account, optionally into a named audience segment; low-risk external mutation, no approval required

## Security

- read risk: external tinyEmail API read of subscriber, list, and campaign data
- write risk: external tinyEmail API mutation: creates or upserts a subscriber (customer) record, optionally assigning it to a named audience segment
- approval: reverse ETL plan approval required before writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect tinyemail
```

### Inspect as structured JSON

```bash
pm connectors inspect tinyemail --json
```

## Agent Rules

- Run pm connectors inspect tinyemail before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
