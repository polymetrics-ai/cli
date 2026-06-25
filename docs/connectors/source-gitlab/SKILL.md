---
name: pm-source-gitlab
description: Gitlab connector knowledge and safe action guide.
---

# pm-source-gitlab

## Purpose

Gitlab catalog connector for https://docs.airbyte.com/integrations/sources/gitlab. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-gitlab:4.4.31 (metadata only; not executed)

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

- API reference: https://docs.gitlab.com/ee/api/rest/
- Future REST API deprecations and removals: https://docs.gitlab.com/ee/api/rest/deprecations.html
- GitLab API OpenAPI specification: https://docs.gitlab.com/ee/api/openapi/openapi.yaml
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/gitlab

## Configuration

- api_url (string): Please enter your basic URL from GitLab instance.
- credentials (object) required
- groups (string): [DEPRECATED] Space-delimited list of groups. e.g. airbyte.io.
- groups_list (array): List of groups. e.g. airbyte.io.
- num_workers (integer): Number of concurrent threads for syncing. Higher values can speed up syncs but may hit rate limits. Adjust based on your GitLab instance rate limits.
- projects (string): [DEPRECATED] Space-delimited list of projects. e.g. airbyte.io/documentation meltano/tap-gitlab.
- projects_list (array): Space-delimited list of projects. e.g. airbyte.io/documentation meltano/tap-gitlab.
- start_date (string): The date from which you'd like to replicate data for GitLab API, in the format YYYY-MM-DDT00:00:00Z. Optional. If not set, all data will be replicated. All data generated after ...
- secret fields: credentials.access_token, credentials.client_id, credentials.client_secret, credentials.refresh_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/gitlab

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-gitlab
```

### Inspect as JSON

```bash
pm connectors inspect source-gitlab --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Gitlab documentation](https://docs.airbyte.com/integrations/sources/gitlab)
