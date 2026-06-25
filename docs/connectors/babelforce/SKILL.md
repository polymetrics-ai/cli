---
name: pm-babelforce
description: Babelforce connector knowledge and safe action guide.
---

# pm-babelforce

## Purpose

Reads Babelforce call reporting, recordings, numbers, and users through the Babelforce v2 REST API.

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
pm connectors inspect babelforce
```

### Inspect as structured JSON

```bash
pm connectors inspect babelforce --json
```

## Agent Rules

- Run pm connectors inspect babelforce before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

