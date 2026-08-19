---
name: pm-mailosaur
description: Mailosaur connector knowledge and safe action guide.
---

# pm-mailosaur

## Purpose

Reads Mailosaur virtual servers, message summaries, and account usage transactions through the Mailosaur REST API.

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
- items_per_page
- mode
- received_after
- server
- username
- password (secret) (required)

## ETL Streams

- servers:
  - primary key: id
  - fields: id(string), messages(integer), name(string), users(array)
- messages:
  - primary key: id
  - cursor: received
  - fields: bcc(array), cc(array), from(array), id(string), received(string), server(string), subject(string), to(array), type(string)
- transactions:
  - primary key: timestamp
  - cursor: timestamp
  - fields: email(integer), previews(integer), sms(integer), timestamp(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Mailosaur API read of virtual-server, message-summary, and usage-transaction data
- approval: none; read-only source connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect mailosaur
```

### Inspect as structured JSON

```bash
pm connectors inspect mailosaur --json
```

## Agent Rules

- Run pm connectors inspect mailosaur before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
