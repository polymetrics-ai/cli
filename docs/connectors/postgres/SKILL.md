---
name: pm-postgres
description: PostgreSQL connector knowledge and safe action guide.
---

# pm-postgres

## Purpose

Reads PostgreSQL tables: dynamically discovers schemas/columns from PostgreSQL system catalogs, snapshots tables, and supports cursor-incremental reads on a configurable cursor column. Read-only source.

## Icon

- id: postgresql
- asset: icons/postgresql.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://www.postgresql.org/docs/current/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: database

## Authentication

- password: Live connections require password authentication; peer/socket and client-certificate modes, including ambient certificates, are unsupported.
  - config: host, database, username
  - secrets: password
  - supports: read=true write=false

## Configuration

- cdc_publication
- cursor_field
- database (required)
- host (required)
- mode
- port
- read_limit
- schema
- sslmode
- sslrootcert
- sslservername
- username (required)
- password (secret) (required when mode is not fixture): Fixture mode does not open a source connection.

## Security

- read risk: low
- write risk: n/a (read-only source)
- approval: none required for read-only sync
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
