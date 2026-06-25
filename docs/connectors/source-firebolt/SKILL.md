---
name: pm-source-firebolt
description: Firebolt connector knowledge and safe action guide.
---

# pm-source-firebolt

## Purpose

Firebolt catalog connector for https://docs.airbyte.com/integrations/sources/firebolt. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: database_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-firebolt:2.0.39 (metadata only; not executed)

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

- Firebolt API reference: https://docs.firebolt.io/godocs/Overview/api-reference.html
- Firebolt authentication: https://docs.firebolt.io/godocs/Guides/managing-your-organization/service-accounts.html
- Firebolt Status: https://status.firebolt.io/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/firebolt

## Configuration

- account (string) required: Firebolt account to login.
- client_id (string) required: Firebolt service account ID.
- client_secret (string) required secret: Firebolt secret, corresponding to the service account ID.
- database (string) required: The database to connect to.
- engine (string) required: Engine name to connect to.
- host (string): The host name of your Firebolt database.
- secret fields: client_secret

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/firebolt

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-firebolt
```

### Inspect as JSON

```bash
pm connectors inspect source-firebolt --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Firebolt documentation](https://docs.airbyte.com/integrations/sources/firebolt)
