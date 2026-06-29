---
name: pm-source-okta
description: Okta connector knowledge and safe action guide.
---

# pm-source-okta

## Purpose

Okta catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/okta.svg
- source: official
- review_status: official_verified
- review_url: https://developer.okta.com/docs/reference/

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

- Okta documentation: https://developer.okta.com/docs/reference/

## Configuration

- credentials (object)
- domain (string): manual intervention needed
- start_date (string): UTC date and time in the format YYYY-MM-DDTHH:MM:SSZ. Any data before this date will not be replicated.
- secret fields: credentials.api_token, credentials.client_id, credentials.client_secret, credentials.key_id, credentials.private_key, credentials.refresh_token

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
pm connectors inspect source-okta
```

### Inspect as JSON

```bash
pm connectors inspect source-okta --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Okta documentation](https://developer.okta.com/docs/reference/)
