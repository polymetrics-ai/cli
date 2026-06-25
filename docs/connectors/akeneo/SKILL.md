---
name: pm-akeneo
description: Akeneo connector knowledge and safe action guide.
---

# pm-akeneo

## Purpose

Reads Akeneo PIM products, categories, families, attributes, and channels through the Akeneo REST API (OAuth2 password grant).

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
pm connectors inspect akeneo
```

### Inspect as structured JSON

```bash
pm connectors inspect akeneo --json
```

## Agent Rules

- Run pm connectors inspect akeneo before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

