# ADR-0001: Keep Sync Semantics in the ETL Engine

## Status

Accepted for this phase.

## Decision

PM will keep sync-mode behavior in the app-level ETL engine and local warehouse store, not inside individual connectors.

## Context

Airbyte separates source read mode from destination write mode. Connectors advertise capabilities and emit records; destination operations materialize final state.

## Consequences

- Connector implementations remain simpler.
- Sync semantics can be tested without live external APIs.
- PostgreSQL-backed storage can later reuse the same mode model and state contract.

