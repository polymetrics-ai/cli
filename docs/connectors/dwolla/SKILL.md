---
name: pm-dwolla
description: Dwolla connector knowledge and safe action guide.
---

# pm-dwolla

## Purpose

Reads Dwolla customers, events, exchange partners, and business classifications via the Dwolla HAL+JSON REST API using OAuth2 client-credentials.

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
pm connectors inspect dwolla
```

### Inspect as structured JSON

```bash
pm connectors inspect dwolla --json
```

## Agent Rules

- Run pm connectors inspect dwolla before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

