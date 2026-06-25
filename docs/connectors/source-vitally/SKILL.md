---
name: pm-source-vitally
description: Vitally connector knowledge and safe action guide.
---

# pm-source-vitally

## Purpose

Vitally catalog connector for https://docs.airbyte.com/integrations/sources/vitally. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-vitally:0.4.3 (metadata only; not executed)

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

- Vitally API documentation: https://docs.vitally.io/pushing-data-to-vitally/rest-api
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/vitally

## Configuration

- basic_auth_header (string) secret: Basic Auth Header
- domain (string) required: Provide only the subdomain part, like https://{your-custom-subdomain}.rest.vitally.io/. Keep empty if you don't have a subdomain.
- secret_token (string) required: sk_live_secret_token
- status (string) required: Status of the Vitally accounts. One of the following values; active, churned, activeOrChurned.
- secret fields: basic_auth_header

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/vitally

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-vitally
```

### Inspect as JSON

```bash
pm connectors inspect source-vitally --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Vitally documentation](https://docs.airbyte.com/integrations/sources/vitally)
