---
name: pm-chift
description: Chift connector knowledge and safe action guide.
---

# pm-chift

## Purpose

Reads and writes Chift consumers, connections, syncs, integrations, datastores, and webhook event definitions through the Chift REST API using a session-token (client credentials) exchange.

## Icon

- id: chift
- asset: icons/chift.svg
- source: official
- review_status: official_verified
- review_url: https://docs.chift.eu/docs/introduction/welcome

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- page_size
- account_id (secret) (required)
- client_id (secret) (required)
- client_secret (secret) (required)

## ETL Streams

- consumers:
  - primary key: consumerid
  - fields: active(boolean), consumerid(string), created_on(string), email(string), name(string), phone(string), redirect_url(string)
- connections:
  - primary key: connectionid
  - fields: api(string), connectionid(string), consumerid(string), created_on(string), name(string), status(string)
- syncs:
  - primary key: syncid
  - fields: consumerid(string), created_on(string), name(string), status(string), syncid(string), updated_on(string)
- integrations:
  - primary key: integrationid
  - fields: api(string), description(string), icon_url(string), integrationid(integer), local_agent(boolean), logo_url(string), name(string), status(string)
- datastores:
  - primary key: id
  - fields: id(string), name(string), status(string)
- webhook_definitions:
  - primary key: event, api
  - fields: api(string), event(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_consumer:
  - endpoint: POST /consumers
  - required fields: name
  - risk: external mutation; approval required
- update_consumer:
  - endpoint: PATCH /consumers/{{ record.consumerid }}
  - required fields: consumerid
  - risk: external mutation; approval required
- delete_consumer:
  - endpoint: DELETE /consumers/{{ record.consumerid }}
  - required fields: consumerid
  - risk: irreversible external deletion; approval required

## Security

- read risk: external Chift API read of consumer/connection/sync/integration/datastore/webhook-definition metadata
- write risk: external mutation of Chift consumer records (create/update/delete); approval required
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect chift
```

### Inspect as structured JSON

```bash
pm connectors inspect chift --json
```

## Agent Rules

- Run pm connectors inspect chift before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
