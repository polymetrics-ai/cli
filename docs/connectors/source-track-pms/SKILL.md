---
name: pm-source-track-pms
description: Track PMS connector knowledge and safe action guide.
---

# pm-source-track-pms

## Purpose

Track PMS catalog connector for https://docs.airbyte.com/integrations/sources/track-pms. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-track-pms:4.3.7 (metadata only; not executed)

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

- Track PMS API: https://www.trackhs.com/api
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/track-pms

## Configuration

- api_key (string) required
- api_secret (string) secret
- customer_domain (string) required
- secret fields: api_secret

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/track-pms

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-track-pms
```

### Inspect as JSON

```bash
pm connectors inspect source-track-pms --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Track PMS documentation](https://docs.airbyte.com/integrations/sources/track-pms)
