---
name: pm-pinterest
description: Pinterest connector knowledge and safe action guide.
---

# pm-pinterest

## Purpose

Reads Pinterest ad accounts, boards, campaigns, ad groups, and audiences through the Pinterest API v5 (OAuth2 refresh-token auth). Read-only.

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
pm connectors inspect pinterest
```

### Inspect as structured JSON

```bash
pm connectors inspect pinterest --json
```

## Agent Rules

- Run pm connectors inspect pinterest before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

