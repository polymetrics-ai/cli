# Overview

Freightview is a Tier-2 quarantine migration of `internal/connectors/freightview`. The bundle mirrors the legacy catalog stream names, primary keys, cursor fields, and field list; the runtime read/check behavior is owned by `internal/connectors/hooks/freightview` during the pre-cutover period.

## Auth setup

Use the same configuration and secret names accepted by the legacy `freightview` connector. Secret-shaped fields are marked with `x-secret` in `spec.json`; the hook delegates to the legacy connector so credential handling remains unchanged and secret values are never logged by the bundle.

## Streams notes

The declared streams are static shadows used for schema, catalog, and surface validation. The Tier-2 hook handles reads and checks by calling the legacy connector, preserving the existing request shape, pagination behavior, record mapping, and fixture mode. The declarative paths under `/__legacy_hook/` are not live API endpoints.

## Write actions & risks

None. This migration preserves the legacy read-only surface and does not add reverse-ETL actions.

## Known limits

- This is a quarantine bridge: hook code currently depends on `internal/connectors/freightview` staying present until the wave 6 cutover replaces or absorbs the delegated behavior.
- Dynamic conformance replay is skipped because the declarative shadow path does not model the connector-specific auth, request body, pagination, or record explosion that caused quarantine; hook unit tests and legacy connector tests are the behavioral proof for this bridge.
