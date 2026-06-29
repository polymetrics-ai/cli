---
name: pm-source-grafana
description: Grafana connector knowledge and safe action guide.
---

# pm-source-grafana

## Purpose

Grafana catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/grafana.svg
- source: official
- review_status: official_verified
- review_url: https://grafana.com/docs/grafana/latest/developer-resources/api-reference/http-api/api-legacy/

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.

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

- Grafana documentation: https://grafana.com/docs/grafana/latest/developer-resources/api-reference/http-api/api-legacy/

## Configuration

- api_key (string) required secret: Grafana API Key or Service Account Token
- url (string) required: The URL of your Grafana instance (e.g., https://your-grafana.grafana.net)
- secret fields: api_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-grafana
```

### Inspect as JSON

```bash
pm connectors inspect source-grafana --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Grafana documentation](https://grafana.com/docs/grafana/latest/developer-resources/api-reference/http-api/api-legacy/)
