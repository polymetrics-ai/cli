---
name: pm-destination-deepset
description: Deepset connector knowledge and safe action guide.
---

# pm-destination-deepset

## Purpose

Deepset catalog connector for https://docs.airbyte.com/integrations/destinations/deepset. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: destination
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: destination_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/destination-deepset:0.1.8 (metadata only; not executed)

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

- family: destination_writer
- priority_wave: 3
- etl_operations: catalog, check, write_append, write_dedup, write_overwrite
- reverse_etl_operations: none until native write conformance passes
- conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

## Official Application Documentation

- No upstream application documentation URL was listed in the imported connector registry.
- Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/deepset

## Configuration

- api_key (string) required secret: Your deepset cloud API key
- base_url (string): URL of deepset Cloud API (e.g. https://api.cloud.deepset.ai, https://api.us.deepset.ai, etc). Defaults to https://api.cloud.deepset.ai.
- retries (number): Number of times to retry an action before giving up.
- workspace (string) required: Name of workspace to which to sync the data.
- secret fields: api_key

## Sync Modes

- supported sync modes: append, append_dedup, overwrite
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/destinations/deepset

## Commands

### Inspect catalog entry

```bash
pm connectors inspect destination-deepset
```

### Inspect as JSON

```bash
pm connectors inspect destination-deepset --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Deepset documentation](https://docs.airbyte.com/integrations/destinations/deepset)
