---
name: pm-xero
description: Xero connector knowledge and safe action guide.
---

# pm-xero

## Purpose

Reads Xero accounting data (invoices, contacts, accounts, bank transactions, items, and payments) from the Xero Accounting API.

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
pm connectors inspect xero
```

### Inspect as structured JSON

```bash
pm connectors inspect xero --json
```

## Agent Rules

- Run pm connectors inspect xero before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

