---
name: pm-recharge
description: Recharge connector knowledge and safe action guide.
---

# pm-recharge

## Purpose

Reads Recharge customers, subscriptions, and orders through the Recharge REST API.

## Icon

- id: recharge
- asset: icons/recharge.svg
- source: official
- review_status: official_verified
- review_url: https://docs.getrecharge.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- api_version
- base_url
- mode
- access_token (secret) (required)

## ETL Streams

- customers:
  - primary key: id
  - fields: created_at(string), email(string), id(integer), updated_at(string)
- subscriptions:
  - primary key: id
  - fields: created_at(string), customer_id(integer), id(integer), status(string), updated_at(string)
- orders:
  - primary key: id
  - fields: created_at(string), customer_id(integer), id(integer), status(string), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Recharge API read of customer, subscription, and order data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Recharge's declared typed write actions.
- Usage: pm recharge <command> [flags]

## Sync Transport

- Source transport: declared
- Destination transport: unsupported
- A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.
- Source executor: declarative_api/declarative_stream_source

## Commands

### Inspect as a manual

```bash
pm connectors inspect recharge
```

### Inspect as structured JSON

```bash
pm connectors inspect recharge --json
```

## Agent Rules

- Run pm connectors inspect recharge before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
