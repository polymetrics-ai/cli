---
name: pm-source-surveycto
description: SurveyCTO connector knowledge and safe action guide.
---

# pm-source-surveycto

## Purpose

SurveyCTO catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/surveycto.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.surveycto.com/05-exporting-and-publishing-data/02-api-access/01.api-access.html

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: native_go
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

- family: custom_go_port
- priority_wave: 3
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- SurveyCTO API documentation: https://docs.surveycto.com/05-exporting-and-publishing-data/02-api-access/01.api-access.html

## Configuration

- form_id (array) required: Unique identifier for one of your forms
- password (string) required secret: Password to authenticate into the SurveyCTO server
- server_name (string) required: The name of the SurveryCTO server
- start_date (string): initial date for survey cto
- username (string) required: Username to authenticate into the SurveyCTO server
- secret fields: password

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
pm connectors inspect source-surveycto
```

### Inspect as JSON

```bash
pm connectors inspect source-surveycto --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [SurveyCTO API documentation](https://docs.surveycto.com/05-exporting-and-publishing-data/02-api-access/01.api-access.html)
