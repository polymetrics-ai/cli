---
name: pm-glassfrog
description: GlassFrog connector knowledge and safe action guide.
---

# pm-glassfrog

## Purpose

Reads GlassFrog circles, roles, people, projects, and assignments through the GlassFrog API v3 (read-only full-refresh source).

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
pm connectors inspect glassfrog
```

### Inspect as structured JSON

```bash
pm connectors inspect glassfrog --json
```

## Agent Rules

- Run pm connectors inspect glassfrog before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

