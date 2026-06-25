---
name: pm-source-braze
description: Braze connector knowledge and safe action guide.
---

# pm-source-braze

## Purpose

Braze catalog connector for https://docs.airbyte.com/integrations/sources/braze. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-braze:0.4.20 (metadata only; not executed)

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

- Braze REST API reference: https://www.braze.com/docs/api/home
- Braze API authentication: https://www.braze.com/docs/api/basics/
- Braze API rate limits: https://www.braze.com/docs/api/api_limits/
- Braze Status: https://status.braze.com/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/braze

## Configuration

- api_key (string) required secret: Braze REST API key
- start_date (string) required: Rows after this date will be synced
- url (string) required: Braze REST API endpoint
- secret fields: api_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/braze

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-braze
```

### Inspect as JSON

```bash
pm connectors inspect source-braze --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Braze documentation](https://docs.airbyte.com/integrations/sources/braze)
