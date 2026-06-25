---
name: pm-source-pocket
description: Pocket connector knowledge and safe action guide.
---

# pm-source-pocket

## Purpose

Pocket catalog connector for https://docs.airbyte.com/integrations/sources/pocket. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-pocket:0.2.37 (metadata only; not executed)

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
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/pocket

## Configuration

- access_token (string) required secret: The user's Pocket access token.
- consumer_key (string) required secret: Your application's Consumer Key.
- content_type (string): Select the content type of the items to retrieve.
- detail_type (string): Select the granularity of the information about each item.
- domain (string): Only return items from a particular `domain`.
- favorite (boolean): Retrieve only favorited items.
- search (string): Only return items whose title or url contain the `search` string.
- since (string): Only return items modified since the given timestamp.
- sort (string): Sort retrieved items by the given criteria.
- state (string): Select the state of the items to retrieve.
- tag (string): Return only items tagged with this tag name. Use _untagged_ for retrieving only untagged items.
- secret fields: access_token, consumer_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/pocket

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-pocket
```

### Inspect as JSON

```bash
pm connectors inspect source-pocket --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Pocket documentation](https://docs.airbyte.com/integrations/sources/pocket)
