---
name: pm-source-zenloop
description: Zenloop connector knowledge and safe action guide.
---

# pm-source-zenloop

## Purpose

Zenloop catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/zenloop.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.zenloop.com/reference

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: beta
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: native_go
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

- family: custom_go_port
- priority_wave: 2
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Zenloop API documentation: https://docs.zenloop.com/reference

## Configuration

- api_token (string) required secret: Zenloop API Token. You can get the API token in settings page <a href="https://app.zenloop.com/settings/api">here</a>
- date_from (string): Zenloop date_from. Format: 2021-10-24T03:30:30Z or 2021-10-24. Leave empty if only data from current data should be synced
- survey_group_id (string) secret: Zenloop Survey Group ID. Can be found by pulling All Survey Groups via SurveyGroups stream. Leave empty to pull answers from all survey groups
- survey_id (string) secret: Zenloop Survey ID. Can be found <a href="https://app.zenloop.com/settings/api">here</a>. Leave empty to pull answers from all surveys
- secret fields: api_token, survey_group_id, survey_id

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
pm connectors inspect source-zenloop
```

### Inspect as JSON

```bash
pm connectors inspect source-zenloop --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Zenloop API documentation](https://docs.zenloop.com/reference)
