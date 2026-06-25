---
name: pm-source-webflow
description: Webflow connector knowledge and safe action guide.
---

# pm-source-webflow

## Purpose

Webflow catalog connector for https://docs.airbyte.com/integrations/sources/webflow. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: native_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-webflow:0.1.46 (metadata only; not executed)

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

- Webflow Data API: https://developers.webflow.com/data/reference
- Webflow authentication: https://developers.webflow.com/data/docs/getting-started
- Webflow API Changelog: https://developers.webflow.com/data/changelog
- Webflow v1 API Deprecation Notice: https://developers.webflow.com/data/docs/webflow-v1-api-deprecation-notice
- Webflow rate limits: https://developers.webflow.com/data/docs/rate-limits
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/webflow

## Configuration

- accept_version (string): The version of the Webflow API to use. See https://developers.webflow.com/#versioning
- api_key (string) required secret: The API token for authenticating to Webflow. See https://university.webflow.com/lesson/intro-to-the-webflow-api
- site_id (string) required: The id of the Webflow site you are requesting data from. See https://developers.webflow.com/#sites
- secret fields: api_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/webflow

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-webflow
```

### Inspect as JSON

```bash
pm connectors inspect source-webflow --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Webflow documentation](https://docs.airbyte.com/integrations/sources/webflow)
