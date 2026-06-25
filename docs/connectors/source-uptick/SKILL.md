---
name: pm-source-uptick
description: Uptick connector knowledge and safe action guide.
---

# pm-source-uptick

## Purpose

Uptick catalog connector for https://docs.airbyte.com/integrations/sources/uptick. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-uptick:0.5.13 (metadata only; not executed)

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
- priority_wave: 1
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Uptick API documentation: https://developer.uptick.com/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/uptick

## Configuration

- base_url (string) required: eg. https://demo-fire.onuptick.com (no trailing slash)
- client_id (string) required secret
- client_secret (string) required secret
- password (string) required secret
- username (string) required
- secret fields: client_id, client_secret, password

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/uptick

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-uptick
```

### Inspect as JSON

```bash
pm connectors inspect source-uptick --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Uptick documentation](https://docs.airbyte.com/integrations/sources/uptick)
