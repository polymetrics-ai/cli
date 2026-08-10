---
name: pm-mysql
description: MySQL connector knowledge and safe action guide.
---

# pm-mysql

## Purpose

Native MySQL source connector for wire-protocol checks, dynamic schemas, and bounded reads. Read-only source.

## Icon

- id: mysql
- asset: icons/mysql.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://dev.mysql.com/doc/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: database

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- cursor_field
- database (required)
- host (required)
- page_size
- port
- read_limit
- sslmode
- sslrootcert
- sslservername
- username (required)
- password (secret)

## Security

- read risk: read-only MySQL wire-protocol queries against the configured database
- write risk: n/a (read-only source)
- approval: none required for read-only sync
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect mysql
```

### Inspect as structured JSON

```bash
pm connectors inspect mysql --json
```

## Agent Rules

- Run pm connectors inspect mysql before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
