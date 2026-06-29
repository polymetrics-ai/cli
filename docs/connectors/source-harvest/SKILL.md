---
name: pm-source-harvest
description: Harvest connector knowledge and safe action guide.
---

# pm-source-harvest

## Purpose

Harvest catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/harvest.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://www.harveststatus.com/

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

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
- priority_wave: 1
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Systems status: https://www.harveststatus.com/
- overview: https://help.getharvest.com/api-v2/

## Configuration

- account_id (string) required secret: Harvest account ID. Required for all Harvest requests in pair with Personal Access Token
- credentials (object): Choose how to authenticate to Harvest.
- num_workers (integer): Number of concurrent threads for syncing. Higher values can speed up syncs but may increase API rate limit usage. Adjust based on your Harvest API plan.
- replication_end_date (string): UTC date and time in the format 2017-01-25T00:00:00Z. Any data after this date will not be replicated.
- replication_start_date (string) required: UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated.
- secret fields: account_id, credentials.api_token, credentials.client_secret, credentials.refresh_token

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
pm connectors inspect source-harvest
```

### Inspect as JSON

```bash
pm connectors inspect source-harvest --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Systems status](https://www.harveststatus.com/)
- [overview](https://help.getharvest.com/api-v2/)
