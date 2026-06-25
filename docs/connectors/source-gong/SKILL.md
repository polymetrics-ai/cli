---
name: pm-source-gong
description: Gong connector knowledge and safe action guide.
---

# pm-source-gong

## Purpose

Gong catalog connector for https://docs.airbyte.com/integrations/sources/gong. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-gong:1.2.7 (metadata only; not executed)

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

- Gong API reference: https://us-66463.app.gong.io/settings/api/documentation
- Gong authentication: https://us-66463.app.gong.io/settings/api/documentation#overview
- Gong rate limits: https://us-66463.app.gong.io/settings/api/documentation#rate-limits
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/gong

## Configuration

- credentials (object) required: Choose how to authenticate to Gong.
- num_workers (integer): Number of concurrent threads for syncing. Higher values can speed up syncs but may increase API rate limit usage. Adjust based on your Gong API plan; the default of 4 is tuned t...
- start_date (string): The date from which to list calls, in the ISO-8601 format; if not specified, the calls start with the earliest recorded call. For web-conference calls recorded by Gong, the date...
- secret fields: credentials.access_key, credentials.access_key_secret, credentials.access_token, credentials.client_secret, credentials.refresh_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/gong

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-gong
```

### Inspect as JSON

```bash
pm connectors inspect source-gong --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Gong documentation](https://docs.airbyte.com/integrations/sources/gong)
