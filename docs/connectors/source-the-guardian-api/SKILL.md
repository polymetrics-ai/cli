---
name: pm-source-the-guardian-api
description: The Guardian API connector knowledge and safe action guide.
---

# pm-source-the-guardian-api

## Purpose

The Guardian API catalog connector for https://docs.airbyte.com/integrations/sources/the-guardian-api. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-the-guardian-api:0.2.26 (metadata only; not executed)

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

- The Guardian Open Platform: https://open-platform.theguardian.com/documentation/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/the-guardian-api

## Configuration

- api_key (string) required secret: Your API Key. See <a href="https://open-platform.theguardian.com/access/">here</a>. The key is case sensitive.
- end_date (string): (Optional) Use this to set the maximum date (YYYY-MM-DD) of the results. Results newer than the end_date will not be shown. Default is set to the current date (today) for increm...
- query (string): (Optional) The query (q) parameter filters the results to only those that include that search term. The q parameter supports AND, OR and NOT operators.
- section (string): (Optional) Use this to filter the results by a particular section. See <a href="https://content.guardianapis.com/sections?api-key=test">here</a> for a list of all sections, and ...
- start_date (string) required: Use this to set the minimum date (YYYY-MM-DD) of the results. Results older than the start_date will not be shown.
- tag (string): (Optional) A tag is a piece of data that is used by The Guardian to categorise content. Use this parameter to filter results by showing only the ones matching the entered tag. S...
- secret fields: api_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/the-guardian-api

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-the-guardian-api
```

### Inspect as JSON

```bash
pm connectors inspect source-the-guardian-api --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [The Guardian API documentation](https://docs.airbyte.com/integrations/sources/the-guardian-api)
