---
name: pm-source-commcare
description: Commcare connector knowledge and safe action guide.
---

# pm-source-commcare

## Purpose

Commcare catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/commcare.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://confluence.dimagi.com/display/commcarepublic/CommCare+HQ+APIs

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

- CommCare API reference: https://confluence.dimagi.com/display/commcarepublic/CommCare+HQ+APIs
- CommCare authentication: https://confluence.dimagi.com/display/commcarepublic/Authentication

## Configuration

- api_key (string) required secret: Commcare API Key
- app_id (string) required secret: The Application ID we are interested in
- project_space (string): Project Space for commcare
- start_date (string) required: UTC date and time in the format 2017-01-25T00:00:00Z. Only records after this date will be replicated.
- secret fields: api_key, app_id

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
pm connectors inspect source-commcare
```

### Inspect as JSON

```bash
pm connectors inspect source-commcare --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [CommCare API reference](https://confluence.dimagi.com/display/commcarepublic/CommCare+HQ+APIs)
- [CommCare authentication](https://confluence.dimagi.com/display/commcarepublic/Authentication)
