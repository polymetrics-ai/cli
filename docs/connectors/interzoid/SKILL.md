---
name: pm-interzoid
description: Interzoid connector knowledge and safe action guide.
---

# pm-interzoid

## Purpose

Reads Interzoid data-matching lookups: company-name, individual-name, and street-address similarity keys, plus organization-name standardization, via the Interzoid REST API.

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
pm connectors inspect interzoid
```

### Inspect as structured JSON

```bash
pm connectors inspect interzoid --json
```

## Agent Rules

- Run pm connectors inspect interzoid before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

