---
name: pm-less-annoying-crm
description: Less Annoying CRM connector knowledge and safe action guide.
---

# pm-less-annoying-crm

## Purpose

Reads Less Annoying CRM users, contacts, tasks, notes, and events through the Less Annoying CRM v2 API.

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
pm connectors inspect less-annoying-crm
```

### Inspect as structured JSON

```bash
pm connectors inspect less-annoying-crm --json
```

## Agent Rules

- Run pm connectors inspect less-annoying-crm before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

