---
name: pm-source-wordpress
description: Wordpress connector knowledge and safe action guide.
---

# pm-source-wordpress

## Purpose

Wordpress catalog connector for https://docs.airbyte.com/integrations/sources/wordpress. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-wordpress:0.0.55 (metadata only; not executed)

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

- WordPress REST API: https://developer.wordpress.org/rest-api/
- WordPress authentication: https://developer.wordpress.org/rest-api/using-the-rest-api/authentication/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/wordpress

## Configuration

- domain (string) required: The domain of the WordPress site. Example: my-wordpress-website.host.com
- lookback_window (integer): Lookback window in hours for incremental streams (editor_blocks, comments, pages, media). Specifies how many hours of previously synced data to re-fetch on each sync to prevent ...
- password (string) required secret: Placeholder for basic HTTP auth password - should be set to empty string
- start_date (string) required: Minimal Date to Retrieve Records when stream allow incremental.
- username (string) required secret: Placeholder for basic HTTP auth username - should be set to empty string
- secret fields: password, username

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/wordpress

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-wordpress
```

### Inspect as JSON

```bash
pm connectors inspect source-wordpress --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Wordpress documentation](https://docs.airbyte.com/integrations/sources/wordpress)
