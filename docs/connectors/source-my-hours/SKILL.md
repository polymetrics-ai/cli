---
name: pm-source-my-hours
description: My Hours connector knowledge and safe action guide.
---

# pm-source-my-hours

## Purpose

My Hours catalog connector for https://docs.airbyte.com/integrations/sources/my-hours. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-my-hours:0.3.46 (metadata only; not executed)

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

- My Hours API documentation: https://myhours.com/api
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/my-hours

## Configuration

- email (string) required: Your My Hours username
- logs_batch_size (integer): Pagination size used for retrieving logs in days
- password (string) required secret: The password associated to the username
- start_date (string) required: Start date for collecting time logs
- secret fields: password

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/my-hours

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-my-hours
```

### Inspect as JSON

```bash
pm connectors inspect source-my-hours --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [My Hours documentation](https://docs.airbyte.com/integrations/sources/my-hours)
