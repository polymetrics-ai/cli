---
name: pm-source-rocket-chat
description: Rocket.chat connector knowledge and safe action guide.
---

# pm-source-rocket-chat

## Purpose

Rocket.chat catalog connector for https://docs.airbyte.com/integrations/sources/rocket-chat. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-rocket-chat:0.2.25 (metadata only; not executed)

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

- No upstream application documentation URL was listed in the imported connector registry.
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/rocket-chat

## Configuration

- endpoint (string) required: Your rocket.chat instance URL.
- token (string) required secret: Your API Token. See <a href="https://developer.rocket.chat/reference/api/rest-api/endpoints/other-important-endpoints/access-tokens-endpoints">here</a>. The token is case sensit...
- user_id (string) required: Your User Id.
- secret fields: token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/rocket-chat

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-rocket-chat
```

### Inspect as JSON

```bash
pm connectors inspect source-rocket-chat --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Rocket.chat documentation](https://docs.airbyte.com/integrations/sources/rocket-chat)
