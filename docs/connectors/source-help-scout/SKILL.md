---
name: pm-source-help-scout
description: Help Scout connector knowledge and safe action guide.
---

# pm-source-help-scout

## Purpose

Help Scout catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
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

- family: declarative_http_source
- priority_wave: 3
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Help Scout API reference: https://developer.helpscout.com/
- Help Scout authentication: https://developer.helpscout.com/mailbox-api/overview/authentication/
- Help Scout rate limits: https://developer.helpscout.com/mailbox-api/overview/rate-limiting/
- Help Scout Status: https://status.helpscout.com/

## Configuration

- client_id (string) required secret
- client_secret (string) required secret
- start_date (string) required
- secret fields: client_id, client_secret

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
pm connectors inspect source-help-scout
```

### Inspect as JSON

```bash
pm connectors inspect source-help-scout --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Help Scout API reference](https://developer.helpscout.com/)
- [Help Scout authentication](https://developer.helpscout.com/mailbox-api/overview/authentication/)
- [Help Scout rate limits](https://developer.helpscout.com/mailbox-api/overview/rate-limiting/)
- [Help Scout Status](https://status.helpscout.com/)
