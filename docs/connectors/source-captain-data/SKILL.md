---
name: pm-source-captain-data
description: Captain Data connector knowledge and safe action guide.
---

# pm-source-captain-data

## Purpose

Captain Data catalog connector for https://docs.airbyte.com/integrations/sources/captain-data. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-captain-data:0.2.25 (metadata only; not executed)

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

- family: declarative_http_source
- priority_wave: 3
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Captain Data API documentation: https://docs.captaindata.co/api-documentation
- Captain Data authentication: https://docs.captaindata.co/api-documentation/authentication
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/captain-data

## Configuration

- api_key (string) required secret: Your Captain Data project API key.
- project_uid (string) required: Your Captain Data project uuid.
- secret fields: api_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/captain-data

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-captain-data
```

### Inspect as JSON

```bash
pm connectors inspect source-captain-data --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Captain Data documentation](https://docs.airbyte.com/integrations/sources/captain-data)
