---
name: pm-pocket
description: Pocket connector knowledge and safe action guide.
---

# pm-pocket

## Purpose

Reads saved Pocket items through the v3 retrieve API.

## Icon

- id: pocket
- asset: icons/pocket.svg
- source: upstream_registry
- review_status: upstream_seeded

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
pm connectors inspect pocket
```

### Inspect as structured JSON

```bash
pm connectors inspect pocket --json
```

## Agent Rules

- Run pm connectors inspect pocket before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
