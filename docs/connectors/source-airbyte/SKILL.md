---
name: pm-source-airbyte
description: Airbyte connector knowledge and safe action guide.
---

# pm-source-airbyte

## Purpose

Airbyte catalog connector for https://docs.airbyte.com/integrations/sources/airbyte. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-airbyte:0.1.24 (metadata only; not executed)

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

- API reference: https://reference.airbyte.com/reference/getting-started
- Authentication: https://docs.airbyte.com/using-airbyte/getting-started/api
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/airbyte

## Configuration

- client_id (string) required
- client_secret (string) required secret
- host (string): The Host URL of your Self-Managed Deployment (e.x. airbtye.mydomain.com)
- start_date (string) required
- secret fields: client_secret

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/airbyte

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-airbyte
```

### Inspect as JSON

```bash
pm connectors inspect source-airbyte --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Airbyte documentation](https://docs.airbyte.com/integrations/sources/airbyte)
