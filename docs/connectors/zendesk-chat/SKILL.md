---
name: pm-zendesk-chat
description: Zendesk Chat connector knowledge and safe action guide.
---

# pm-zendesk-chat

## Purpose

Reads Zendesk Chat agents, chats, departments, shortcuts, and triggers through the Zendesk Chat REST API v2.

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
pm connectors inspect zendesk-chat
```

### Inspect as structured JSON

```bash
pm connectors inspect zendesk-chat --json
```

## Agent Rules

- Run pm connectors inspect zendesk-chat before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

