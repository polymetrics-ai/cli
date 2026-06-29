---
name: pm-source-instagram
description: Instagram connector knowledge and safe action guide.
---

# pm-source-instagram

## Purpose

Instagram catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/instagram.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.facebook.com/docs/instagram-platform/changelog

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

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
- priority_wave: 1
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Instagram Platform Changelog: https://developers.facebook.com/docs/instagram-platform/changelog
- Release notes: https://developers.facebook.com/docs/instagram-api/changelog

## Configuration

- access_token (string) required secret: manual intervention needed
- client_id (string) secret: The Client ID for your Oauth application
- client_secret (string) secret: The Client Secret for your Oauth application
- num_workers (integer): The number of worker threads to use for the sync.
- start_date (string): The date from which you'd like to replicate data for User Insights, in the format YYYY-MM-DDT00:00:00Z. All data generated after this date will be replicated. If left blank, the...
- secret fields: access_token, client_id, client_secret

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
pm connectors inspect source-instagram
```

### Inspect as JSON

```bash
pm connectors inspect source-instagram --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Instagram Platform Changelog](https://developers.facebook.com/docs/instagram-platform/changelog)
- [Release notes](https://developers.facebook.com/docs/instagram-api/changelog)
