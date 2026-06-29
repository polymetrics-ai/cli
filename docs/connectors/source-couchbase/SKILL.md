---
name: pm-source-couchbase
description: Couchbase connector knowledge and safe action guide.
---

# pm-source-couchbase

## Purpose

Couchbase catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/couchbase.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.couchbase.com/server/current/n1ql/n1ql-language-reference/index.html

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: database_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.

## Runtime Capabilities

- metadata=true
- check=false
- catalog=false
- read=false
- write=false
- query=false
- etl=false
- reverse_etl=false
- unsupported_reason: Native Go port is planned but not enabled; only catalog metadata is available.

## Native Port Plan

- family: database_source
- priority_wave: 3
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, cursor_incremental, docs_skill, query_safety, read_fixture, secret_redaction, spec, state_checkpoint, type_mapping

## Official Application Documentation

- Couchbase SQL++ reference: https://docs.couchbase.com/server/current/n1ql/n1ql-language-reference/index.html
- Couchbase authentication: https://docs.couchbase.com/server/current/learn/security/authentication.html

## Configuration

- bucket (string) required: The name of the bucket to sync data from
- connection_string (string) required: The connection string for the Couchbase server (e.g., couchbase://localhost or couchbases://example.com)
- password (string) required secret: The password to use for authentication
- start_date (string): The date from which you'd like to replicate data for incremental streams, in the format YYYY-MM-DDT00:00:00Z. All data generated after this date will be replicated. If not set, ...
- username (string) required: The username to use for authentication
- secret fields: password

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-couchbase
```

### Inspect as JSON

```bash
pm connectors inspect source-couchbase --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Couchbase SQL++ reference](https://docs.couchbase.com/server/current/n1ql/n1ql-language-reference/index.html)
- [Couchbase authentication](https://docs.couchbase.com/server/current/learn/security/authentication.html)
