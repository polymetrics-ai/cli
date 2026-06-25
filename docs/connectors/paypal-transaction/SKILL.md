---
name: pm-paypal-transaction
description: PayPal Transaction connector knowledge and safe action guide.
---

# pm-paypal-transaction

## Purpose

Reads PayPal transactions, balances, catalog products, and customer disputes through the PayPal REST API using OAuth 2.0 client-credentials auth.

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- No connector-specific config fields.

## Security

- read risk: connector-specific
- write risk: connector-specific
- approval: external mutations require preview and approval
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect paypal-transaction
```

### Inspect as structured JSON

```bash
pm connectors inspect paypal-transaction --json
```

## Agent Rules

- Run pm connectors inspect paypal-transaction before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

