---
name: pm-dropbox-sign
description: Dropbox Sign connector knowledge and safe action guide.
---

# pm-dropbox-sign

## Purpose

Reads Dropbox Sign (HelloSign) signature requests, templates, team members, and account details through the Dropbox Sign REST API.

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
pm connectors inspect dropbox-sign
```

### Inspect as structured JSON

```bash
pm connectors inspect dropbox-sign --json
```

## Agent Rules

- Run pm connectors inspect dropbox-sign before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

