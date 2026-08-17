---
name: pm-commercetools
description: commercetools connector knowledge and safe action guide.
---

# pm-commercetools

## Purpose

Reads commercetools customers, orders, and products through the HTTP API.

## Icon

- asset: icons/commercetools.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.commercetools.com/api/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- project_key
- token_url
- client_id (secret)
- client_secret (secret)

## ETL Streams

- customers:
  - primary key: id
  - cursor: createdAt
  - fields: addresses(), authenticationMode(), createdAt(), customerNumber(), email(), firstName(), id(), isEmailVerified(), lastModifiedAt(), lastName(), version()
- orders:
  - primary key: id
  - cursor: createdAt
  - fields: createdAt(), customerId(), id(), lastModifiedAt(), lineItems(), orderNumber(), orderState(), totalPrice(), version()
- products:
  - primary key: id
  - cursor: createdAt
  - fields: createdAt(), id(), lastModifiedAt(), masterData(), productType(), version()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external commercetools API read of customer, order, and product data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect commercetools
```

### Inspect as structured JSON

```bash
pm connectors inspect commercetools --json
```

## Agent Rules

- Run pm connectors inspect commercetools before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
