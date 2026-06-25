---
name: pm-source-square
description: Square connector knowledge and safe action guide.
---

# pm-source-square

## Purpose

Square catalog connector for https://docs.airbyte.com/integrations/sources/square. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: beta
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-square:1.7.20 (metadata only; not executed)

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

- Square API reference: https://developer.squareup.com/reference/square
- Square authentication: https://developer.squareup.com/docs/build-basics/access-tokens
- Square API Release Notes: https://developer.squareup.com/docs/release-notes
- Square API changelog: https://developer.squareup.com/docs/changelog
- Square rate limits: https://developer.squareup.com/docs/build-basics/api-rate-limits
- Square Status: https://www.issquareup.com/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/square

## Configuration

- credentials (object): Choose how to authenticate to Square.
- include_deleted_objects (boolean): In some streams there is an option to include deleted objects (Items, Categories, Discounts, Taxes)
- is_sandbox (boolean) required: Determines whether to use the sandbox or production environment.
- start_date (string): UTC date in the format YYYY-MM-DD. Any data before this date will not be replicated. If not set, all data will be replicated.
- secret fields: credentials.api_key, credentials.client_id, credentials.client_secret, credentials.refresh_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/square

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-square
```

### Inspect as JSON

```bash
pm connectors inspect source-square --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Square documentation](https://docs.airbyte.com/integrations/sources/square)
