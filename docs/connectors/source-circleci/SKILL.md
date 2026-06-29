---
name: pm-source-circleci
description: Circleci connector knowledge and safe action guide.
---

# pm-source-circleci

## Purpose

Circleci catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

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

- CircleCI API v2 reference: https://circleci.com/docs/api/v2/
- CircleCI authentication: https://circleci.com/docs/api-developers-guide/#authentication
- CircleCI API rate limits: https://circleci.com/docs/api-developers-guide/#rate-limits
- CircleCI Status: https://status.circleci.com/

## Configuration

- api_key (string) required secret
- job_number (string): Job Number of the workflow for `jobs` stream, Auto fetches from `workflow_jobs` stream, if not configured
- org_id (string) required: The org ID found in `https://app.circleci.com/settings/organization/circleci/xxxxx/overview`
- project_id (string) required: Project ID found in the project settings, Visit `https://app.circleci.com/settings/project/circleci/ORG_SLUG/YYYYY`
- start_date (string) required
- workflow_id (array): Workflow ID of a project pipeline, Could be seen in the URL of pipeline build, Example `https://app.circleci.com/pipelines/circleci/55555xxxxxx/7yyyyyyyyxxxxx/2/workflows/WORKFL...
- secret fields: api_key

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
pm connectors inspect source-circleci
```

### Inspect as JSON

```bash
pm connectors inspect source-circleci --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [CircleCI API v2 reference](https://circleci.com/docs/api/v2/)
- [CircleCI authentication](https://circleci.com/docs/api-developers-guide/#authentication)
- [CircleCI API rate limits](https://circleci.com/docs/api-developers-guide/#rate-limits)
- [CircleCI Status](https://status.circleci.com/)
