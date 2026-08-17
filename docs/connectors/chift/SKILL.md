---
name: pm-chift
description: Chift connector knowledge and safe action guide.
---

# pm-chift

## Purpose

Reads and writes Chift consumers, connections, syncs, integrations, datastores, and webhook event definitions through the Chift REST API using a session-token (client credentials) exchange.

## Icon

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
- account_id (secret)
- client_id (secret)
- client_secret (secret)

## ETL Streams

- consumers:
  - primary key: consumerid
  - fields: active(), consumerid(), created_on(), email(), name(), phone(), redirect_url()
- connections:
  - primary key: connectionid
  - fields: api(), connectionid(), consumerid(), created_on(), name(), status()
- syncs:
  - primary key: syncid
  - fields: consumerid(), created_on(), name(), status(), syncid(), updated_on()
- integrations:
  - primary key: integrationid
  - fields: api(), description(), icon_url(), integrationid(), local_agent(), logo_url(), name(), status()
- datastores:
  - primary key: id
  - fields: id(), name(), status()
- webhook_definitions:
  - primary key: event, api
  - fields: api(), event()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_consumer:
  - endpoint: POST /consumers
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
