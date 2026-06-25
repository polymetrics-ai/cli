---
name: pm-source-metabase
description: Metabase connector knowledge and safe action guide.
---

# pm-source-metabase

## Purpose

Metabase catalog connector for https://docs.airbyte.com/integrations/sources/metabase. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: beta
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-metabase:2.1.49 (metadata only; not executed)

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
- priority_wave: 2
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Metabase API reference: https://www.metabase.com/docs/latest/api-documentation
- Metabase authentication: https://www.metabase.com/docs/latest/api/session
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/metabase

## Configuration

- instance_api_url (string) required: URL to your metabase instance API
- password (string) secret
- session_token (string) secret: To generate your session token, you need to run the following command: ``` curl -X POST \ -H "Content-Type: application/json" \ -d '{"username": "person@metabase.com", "password...
- username (string) required
- secret fields: password, session_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/metabase

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-metabase
```

### Inspect as JSON

```bash
pm connectors inspect source-metabase --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Metabase documentation](https://docs.airbyte.com/integrations/sources/metabase)
