---
name: pm-lemlist
description: Lemlist connector knowledge and safe action guide.
---

# pm-lemlist

## Purpose

Reads lemlist team, campaigns, activities, and unsubscribes through the lemlist REST API.

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
pm connectors inspect lemlist
```

### Inspect as structured JSON

```bash
pm connectors inspect lemlist --json
```

## Agent Rules

- Run pm connectors inspect lemlist before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

