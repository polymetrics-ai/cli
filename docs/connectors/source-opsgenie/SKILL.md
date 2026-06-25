---
name: pm-source-opsgenie
description: Opsgenie connector knowledge and safe action guide.
---

# pm-source-opsgenie

## Purpose

Opsgenie catalog connector for https://docs.airbyte.com/integrations/sources/opsgenie. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-opsgenie:0.5.12 (metadata only; not executed)

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

- Opsgenie API reference: https://docs.opsgenie.com/docs/api-overview
- Opsgenie authentication: https://docs.opsgenie.com/docs/api-authentication
- Opsgenie rate limits: https://docs.opsgenie.com/docs/api-rate-limiting
- Opsgenie Status: https://status.opsgenie.com/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/opsgenie

## Configuration

- api_token (string) required secret: API token used to access the Opsgenie platform
- endpoint (string) required: Service endpoint to use for API calls.
- start_date (string): The date from which you'd like to replicate data from Opsgenie in the format of YYYY-MM-DDT00:00:00Z. All data generated after this date will be replicated. Note that it will be...
- secret fields: api_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/opsgenie

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-opsgenie
```

### Inspect as JSON

```bash
pm connectors inspect source-opsgenie --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Opsgenie documentation](https://docs.airbyte.com/integrations/sources/opsgenie)
