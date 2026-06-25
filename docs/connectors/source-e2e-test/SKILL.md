---
name: pm-source-e2e-test
description: E2E Testing connector knowledge and safe action guide.
---

# pm-source-e2e-test

## Purpose

E2E Testing catalog connector for https://docs.airbyte.com/integrations/sources/e2e-test. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: native_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-e2e-test:2.2.2 (metadata only; not executed)

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

- Airbyte E2E test source documentation: https://docs.airbyte.com/integrations/sources/e2e-test
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/e2e-test

## Configuration

- No config schema fields were advertised in the catalog.

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/e2e-test

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-e2e-test
```

### Inspect as JSON

```bash
pm connectors inspect source-e2e-test --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [E2E Testing documentation](https://docs.airbyte.com/integrations/sources/e2e-test)
