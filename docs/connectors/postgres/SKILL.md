---
name: pm-postgres
description: PostgreSQL connector knowledge and safe action guide.
---

# pm-postgres

## Purpose

Reads PostgreSQL tables: discovers schemas/columns from information_schema, snapshots tables, and supports cursor-incremental reads on a configurable cursor column. Read-only source; CDC is a documented stub pending the gated pglogrepl dependency.

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: database

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
pm connectors inspect postgres
```

### Inspect as structured JSON

```bash
pm connectors inspect postgres --json
```

## Agent Rules

- Run pm connectors inspect postgres before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

