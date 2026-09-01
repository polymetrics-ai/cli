---
name: pm-postgres
description: PostgreSQL connector knowledge and safe action guide.
---

# pm-postgres

## Purpose

Reads PostgreSQL tables, supports PostgreSQL 14+ logical-replication CDC into the local warehouse, and delivers managed PostgreSQL targets through the warehouse-mediated transport. Direct Connector.Write remains unavailable; query is not implemented.

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
- write risk: not a generic write capability; the closed managed target transport follows its existing plan, preview, approval, execute boundary
- approval: source reads require none; managed target transport follows its plan, preview, approval, execute boundary
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Sync Transport

- Source transport: declared
- Destination transport: declared
- A declared transport executes only when its named runtime executor and mode are available.
- Source executor: native_database/postgres_polling_watermark
- Destination executor: native_database/postgres_managed_target

## Polling Watermark

- Static declaration status: planned
- Mechanism: polling_watermark is a bounded polling scan, not CDC or change capture.
- Runtime eligibility: this connector constructs an implemented declaration per selected catalog object. Every requested mode still requires runtime preflight for its destination binding, registered native executors, and compatible runtime mode.
- Reason: PostgreSQL binds its cursor type, per-stream cursor column, and unique tie-breaker from the live catalog at run time. The static bundle cannot truthfully name those dynamic fields; the live sync transport constructs and preflights the implemented polling declaration.

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
