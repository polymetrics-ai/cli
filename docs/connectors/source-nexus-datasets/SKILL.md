---
name: pm-source-nexus-datasets
description: Infor Nexus connector knowledge and safe action guide.
---

# pm-source-nexus-datasets

## Purpose

Infor Nexus catalog connector for https://docs.airbyte.com/integrations/sources/nexus-datasets. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-nexus-datasets:0.1.9 (metadata only; not executed)

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

- No upstream application documentation URL was listed in the imported connector registry.
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/nexus-datasets

## Configuration

- access_key_id (string) required: The access key ID for the Infor Nexus API.
- api_key (string) required: The Infor Data API Key.
- base_url (string) required: The base URL of the Infor Nexus datasets API.
- dataset_name (string) required: The name of the dataset to export.
- mode (string): Infor Streaming mode (Full or Incremental). https://developer.infornexus.com/api/version-3.1-api-methods
- secret_key (string) required: The secret key for HMAC authentication.
- user_id (string) required: The user ID for the Infor Nexus API.

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/nexus-datasets

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-nexus-datasets
```

### Inspect as JSON

```bash
pm connectors inspect source-nexus-datasets --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Infor Nexus documentation](https://docs.airbyte.com/integrations/sources/nexus-datasets)
