---
name: pm-source-notion
description: Notion connector knowledge and safe action guide.
---

# pm-source-notion

## Purpose

Notion catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/notion.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.notion.com/reference/changes-by-version

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

- Changes by version: https://developers.notion.com/reference/changes-by-version
- Changelog: https://developers.notion.com/page/changelog

## Configuration

- credentials (object): manual intervention needed
- num_workers (integer): Number of worker threads to use for the sync. Higher values can speed up large syncs but may increase rate-limit pressure against Notion's limit of approximately three requests ...
- start_date (string): UTC date and time in the format YYYY-MM-DDTHH:MM:SS.000Z. During incremental sync, any data generated before this date will not be replicated. If left blank, the start date will...
- secret fields: credentials.access_token, credentials.client_id, credentials.client_secret, credentials.token

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
pm connectors inspect source-notion
```

### Inspect as JSON

```bash
pm connectors inspect source-notion --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Changes by version](https://developers.notion.com/reference/changes-by-version)
- [Changelog](https://developers.notion.com/page/changelog)
