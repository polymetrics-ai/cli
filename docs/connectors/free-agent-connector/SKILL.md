---
name: pm-free-agent-connector
description: FreeAgent connector knowledge and safe action guide.
---

# pm-free-agent-connector

## Purpose

Reads FreeAgent contacts, invoices, bills, projects, and tasks through the FreeAgent v2 REST API using OAuth2 refresh-token authentication.

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
pm connectors inspect free-agent-connector
```

### Inspect as structured JSON

```bash
pm connectors inspect free-agent-connector --json
```

## Agent Rules

- Run pm connectors inspect free-agent-connector before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

