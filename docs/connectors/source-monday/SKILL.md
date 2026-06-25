---
name: pm-source-monday
description: Monday connector knowledge and safe action guide.
---

# pm-source-monday

## Purpose

Monday catalog connector for https://docs.airbyte.com/integrations/sources/monday. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-monday:2.5.11 (metadata only; not executed)

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

- monday.com API reference: https://developer.monday.com/api-reference/docs
- monday.com authentication: https://developer.monday.com/api-reference/docs/authentication
- monday.com rate limits: https://developer.monday.com/api-reference/docs/rate-limits
- monday.com Status: https://status.monday.com/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/monday

## Configuration

- board_ids (array): The IDs of the boards that the Items and Boards streams will extract records from. When left empty, streams will extract records from all boards that exist within the account.
- credentials (object)
- num_workers (integer): The number of worker threads to use for the sync.
- secret fields: credentials.access_token, credentials.api_token, credentials.client_id, credentials.client_secret

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/monday

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-monday
```

### Inspect as JSON

```bash
pm connectors inspect source-monday --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Monday documentation](https://docs.airbyte.com/integrations/sources/monday)
