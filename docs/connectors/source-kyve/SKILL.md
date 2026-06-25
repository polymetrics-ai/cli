---
name: pm-source-kyve
description: KYVE connector knowledge and safe action guide.
---

# pm-source-kyve

## Purpose

KYVE catalog connector for https://docs.airbyte.com/integrations/sources/kyve. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: native_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-kyve:0.2.52 (metadata only; not executed)

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

- family: custom_go_port
- priority_wave: 3
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- KYVE documentation: https://docs.kyve.network/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/kyve

## Configuration

- max_pages (integer): The maximum amount of pages to go trough. Set to 'null' for all pages.
- page_size (integer): The pagesize for pagination, smaller numbers are used in integration tests.
- pool_ids (string) required: The IDs of the KYVE storage pool you want to archive. (Comma separated)
- start_ids (string) required: The start-id defines, from which bundle id the pipeline should start to extract the data. (Comma separated)
- url_base (string) required: URL to the KYVE Chain API.

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/kyve

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-kyve
```

### Inspect as JSON

```bash
pm connectors inspect source-kyve --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [KYVE documentation](https://docs.airbyte.com/integrations/sources/kyve)
