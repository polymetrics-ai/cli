---
name: pm-source-merge
description: Merge connector knowledge and safe action guide.
---

# pm-source-merge

## Purpose

Merge catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/merge.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.merge.dev/api-reference/

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

- Merge API reference: https://docs.merge.dev/api-reference/
- Merge authentication: https://docs.merge.dev/basics/authentication/
- Merge rate limits: https://docs.merge.dev/basics/rate-limits/
- Merge Status: https://status.merge.dev/

## Configuration

- account_token (string) required secret: Link your other integrations with account credentials on accounts section to get account token (ref - https://app.merge.dev/linked-accounts/accounts)
- api_token (string) required secret: API token can be seen at https://app.merge.dev/keys
- start_date (string) required: Date time filter for incremental filter, Specify which date to extract from.
- secret fields: account_token, api_token

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
pm connectors inspect source-merge
```

### Inspect as JSON

```bash
pm connectors inspect source-merge --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Merge API reference](https://docs.merge.dev/api-reference/)
- [Merge authentication](https://docs.merge.dev/basics/authentication/)
- [Merge rate limits](https://docs.merge.dev/basics/rate-limits/)
- [Merge Status](https://status.merge.dev/)
