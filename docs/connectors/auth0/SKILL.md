---
name: pm-auth0
description: Auth0 connector knowledge and safe action guide.
---

# pm-auth0

## Purpose

Reads Auth0 users, clients, connections, roles, and organizations from the Auth0 Management API v2.

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
pm connectors inspect auth0
```

### Inspect as structured JSON

```bash
pm connectors inspect auth0 --json
```

## Agent Rules

- Run pm connectors inspect auth0 before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

