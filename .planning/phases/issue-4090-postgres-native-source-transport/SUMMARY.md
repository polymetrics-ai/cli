---
issue: 4090
status: implementation-complete-pending-inline-review
key_files:
  created:
    - internal/connectors/native/postgres/transport_source.go
  modified:
    - internal/connectors/native/postgres/connector.go
    - internal/connectors/native/postgres/transport_source_test.go
    - internal/connectors/native/postgres/dynamic_catalog_integration_test.go
---

# Summary — Issue #4090

The PostgreSQL native connector now declares a closed `native_database`
bounded-snapshot source in `Definition()`. Its PostgreSQL-local registration
helper resolves that exact declaration without App composition. The executor
validates the closed source request before it opens a pool, uses a read-only
repeatable-read transaction for the existing typed catalog discovery and all
bounded key-ordered reads, and emits a source-bound checkpoint with the typed
catalog fingerprint and PostgreSQL snapshot barrier.

Focused preflight tests prove missing descriptor, wrong-family, and unregistered
executor refusal before source I/O. The typed pagination boundary converts
driver-native UUID and temporal-infinity cursor values into pgx parameters and
refuses non-encodable logical key kinds before issuing the next query. Live
PostgreSQL 16.10 dbtest coverage proves five rows in three pages for
`full_append` and `full_overwrite`, including a UUID, a zone-less timestamp,
an exact JSON integer beyond the `float64` range, and JSON null. It also proves
batch-size-one page two for a `-infinity` timestamp key and a UUID key, with
identity, schema fingerprint, barrier, and dedupe checkpoint emitted by the
registered source.
