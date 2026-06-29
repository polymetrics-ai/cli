---
name: pm-source-devin-ai
description: Devin AI connector knowledge and safe action guide.
---

# pm-source-devin-ai

## Purpose

Devin AI catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/devin-ai.svg
- source: upstream_registry
- review_status: upstream_seeded

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

- Devin AI documentation: https://docs.devin.ai/api-reference/overview

## Configuration

- api_token (string) required secret: Devin API key for authentication (cog_* prefix for service users).
- org_id (string) required: Your Devin organization ID (org_* prefix).
- start_date (string): Optional UTC date-time. Only sessions created on or after this instant are returned for the `sessions`, `sessions_insights`, and `session_messages` streams. Leave empty to fetch...
- secret fields: api_token

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
pm connectors inspect source-devin-ai
```

### Inspect as JSON

```bash
pm connectors inspect source-devin-ai --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Devin AI documentation](https://docs.devin.ai/api-reference/overview)
