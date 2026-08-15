---
name: pm-postgres
description: PostgreSQL connector knowledge and safe action guide.
---

# pm-postgres

## Purpose

Reads PostgreSQL tables, writes warehouse worksets to owned managed PostgreSQL targets, and supports PostgreSQL 14+ logical-replication CDC. Generic SQL and direct connector writes remain unavailable.

## Icon

- id: postgresql
- asset: icons/postgresql.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://www.postgresql.org/docs/current/

## Capabilities

- check=true catalog=true read=true write=true query=false
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
- write risk: high (owned managed targets only)
- approval: warehouse workset and managed-target ownership required for writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Sync Transport

- Source transport: declared
- Destination transport: unsupported
- A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.
- Source executor: native_database/postgres_bounded_snapshot

## Polling Watermark

- Status: planned
- Mechanism: polling_watermark is a bounded polling scan, not CDC or change capture.
- Eligibility: each mode remains blocked until runtime preflight validates the selected catalog object and destination binding, registered native executors, and immutable conformance evidence.
- Reason: no registered native polling source and apply binding has passed object and destination preflight; no polling sync mode is implemented
- No polling source ordering, checkpoint, snapshot, deletion, or rebootstrap behavior is implemented for this connector while the declaration is non-implemented.

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
