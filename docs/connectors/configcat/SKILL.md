---
name: pm-configcat
description: ConfigCat connector knowledge and safe action guide.
---

# pm-configcat

## Purpose

Reads ConfigCat organizations, products, configs, environments, and tags through the ConfigCat Public Management API.

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
pm connectors inspect configcat
```

### Inspect as structured JSON

```bash
pm connectors inspect configcat --json
```

## Agent Rules

- Run pm connectors inspect configcat before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

