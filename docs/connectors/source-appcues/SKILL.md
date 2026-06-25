---
name: pm-source-appcues
description: Appcues connector knowledge and safe action guide.
---

# pm-source-appcues

## Purpose

Appcues catalog connector for https://docs.airbyte.com/integrations/sources/appcues. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-appcues:0.0.50 (metadata only; not executed)

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

- API documentation: https://docs.appcues.com/en_US/api
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/appcues

## Configuration

- account_id (string) required: Account ID of Appcues found in account settings page (https://studio.appcues.com/settings/account)
- password (string) secret
- start_date (string) required
- username (string) required
- secret fields: password

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/appcues

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-appcues
```

### Inspect as JSON

```bash
pm connectors inspect source-appcues --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Appcues documentation](https://docs.airbyte.com/integrations/sources/appcues)
