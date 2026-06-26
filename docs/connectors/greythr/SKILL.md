---
name: pm-greythr
description: greytHR connector knowledge and safe action guide.
---

# pm-greythr

## Purpose

Reads greytHR employees, profiles, work details, bank details, and users via the greytHR REST API using session-token authentication.

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
pm connectors inspect greythr
```

### Inspect as structured JSON

```bash
pm connectors inspect greythr --json
```

## Agent Rules

- Run pm connectors inspect greythr before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

