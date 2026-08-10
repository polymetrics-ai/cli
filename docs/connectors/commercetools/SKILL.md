---
name: pm-commercetools
description: commercetools connector knowledge and safe action guide.
---

# pm-commercetools

## Purpose

Reads commercetools customers, orders, and products through the HTTP API.

## Icon

- id: commercetools
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

- base_url (required)
- mode
- project_key (required)
- token_url (required)
- client_id (secret) (required)
- client_secret (secret) (required)

## ETL Streams

- customers:
  - primary key: id
  - cursor: createdAt
  - fields: addresses(array), authenticationMode(string), createdAt(string), customerNumber(string), email(string), firstName(string), id(string), isEmailVerified(boolean), lastModifiedAt(string), lastName(string), version(integer)
- orders:
  - primary key: id
  - cursor: createdAt
  - fields: createdAt(string), customerId(string), id(string), lastModifiedAt(string), lineItems(array), orderNumber(string), orderState(string), totalPrice(object), version(integer)
- products:
  - primary key: id
  - cursor: createdAt
  - fields: createdAt(string), id(string), lastModifiedAt(string), masterData(object), productType(object), version(integer)

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
