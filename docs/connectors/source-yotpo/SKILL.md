---
name: pm-source-yotpo
description: Yotpo connector knowledge and safe action guide.
---

# pm-source-yotpo

## Purpose

Yotpo catalog connector for https://docs.airbyte.com/integrations/sources/yotpo. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-yotpo:0.2.16 (metadata only; not executed)

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

- Yotpo API documentation: https://apidocs.yotpo.com/reference
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/yotpo

## Configuration

- access_token (string) required secret: Access token recieved as a result of API call to https://api.yotpo.com/oauth/token (Ref- https://apidocs.yotpo.com/reference/yotpo-authentication)
- app_key (string) required: App key found at settings (Ref- https://settings.yotpo.com/#/general_settings)
- email (string) required: Email address registered with yotpo.
- start_date (string) required: Date time filter for incremental filter, Specify which date to extract from.
- secret fields: access_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/yotpo

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-yotpo
```

### Inspect as JSON

```bash
pm connectors inspect source-yotpo --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Yotpo documentation](https://docs.airbyte.com/integrations/sources/yotpo)
