---
name: pm-source-lever-hiring
description: Lever Hiring connector knowledge and safe action guide.
---

# pm-source-lever-hiring

## Purpose

Lever Hiring catalog connector for https://docs.airbyte.com/integrations/sources/lever-hiring. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-lever-hiring:0.4.34 (metadata only; not executed)

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

- Lever API reference: https://hire.lever.co/developer/documentation
- Lever authentication: https://hire.lever.co/developer/documentation#authentication
- Lever rate limits: https://hire.lever.co/developer/documentation#rate-limiting
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/lever-hiring

## Configuration

- credentials (object): Choose how to authenticate to Lever Hiring.
- environment (string): The environment in which you'd like to replicate data for Lever. This is used to determine which Lever API endpoint to use.
- start_date (string) required: UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated. Note that it will be used only in the following incremental streams: comm...
- secret fields: credentials.api_key, credentials.client_secret, credentials.refresh_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/lever-hiring

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-lever-hiring
```

### Inspect as JSON

```bash
pm connectors inspect source-lever-hiring --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Lever Hiring documentation](https://docs.airbyte.com/integrations/sources/lever-hiring)
