---
name: pm-source-looker
description: Looker connector knowledge and safe action guide.
---

# pm-source-looker

## Purpose

Looker catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/looker.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://cloud.google.com/looker/docs/reference/looker-api/latest

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

- Looker API reference: https://cloud.google.com/looker/docs/reference/looker-api/latest
- Looker authentication: https://cloud.google.com/looker/docs/api-auth
- Looker rate limits: https://cloud.google.com/looker/docs/api-rate-limits

## Configuration

- client_id (string) required: manual intervention needed
- client_secret (string) required secret: The Client Secret is second part of an API3 key.
- domain (string) required: manual intervention needed
- run_look_ids (array): The IDs of any Looks to run
- secret fields: client_secret

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
pm connectors inspect source-looker
```

### Inspect as JSON

```bash
pm connectors inspect source-looker --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Looker API reference](https://cloud.google.com/looker/docs/reference/looker-api/latest)
- [Looker authentication](https://cloud.google.com/looker/docs/api-auth)
- [Looker rate limits](https://cloud.google.com/looker/docs/api-rate-limits)
