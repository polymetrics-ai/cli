---
name: pm-source-marketo
description: Marketo connector knowledge and safe action guide.
---

# pm-source-marketo

## Purpose

Marketo catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/marketo.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.marketo.com/rest-api/

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
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
- priority_wave: 1
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Marketo REST API: https://developers.marketo.com/rest-api/
- Marketo authentication: https://developers.marketo.com/rest-api/authentication/
- Marketo rate limits: https://developers.marketo.com/rest-api/marketo-integration-best-practices/#api_limits

## Configuration

- client_id (string) required secret: manual intervention needed
- client_secret (string) required secret: manual intervention needed
- domain_url (string) required secret: manual intervention needed
- start_date (string) required: UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated.
- secret fields: client_id, client_secret, domain_url

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
pm connectors inspect source-marketo
```

### Inspect as JSON

```bash
pm connectors inspect source-marketo --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Marketo REST API](https://developers.marketo.com/rest-api/)
- [Marketo authentication](https://developers.marketo.com/rest-api/authentication/)
- [Marketo rate limits](https://developers.marketo.com/rest-api/marketo-integration-best-practices/#api_limits)
