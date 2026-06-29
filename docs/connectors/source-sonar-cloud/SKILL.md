---
name: pm-source-sonar-cloud
description: Sonar Cloud connector knowledge and safe action guide.
---

# pm-source-sonar-cloud

## Purpose

Sonar Cloud catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/sonarcloud.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://sonarcloud.io/web_api

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

- SonarCloud Web API: https://sonarcloud.io/web_api
- SonarCloud authentication: https://docs.sonarsource.com/sonarcloud/advanced-setup/user-accounts/

## Configuration

- component_keys (array) required: Comma-separated list of component keys.
- end_date (string): To retrieve issues created before the given date (inclusive).
- organization (string) required: Organization key. See <a href="https://docs.sonarcloud.io/appendices/project-information/#project-and-organization-keys">here</a>.
- start_date (string): To retrieve issues created after the given date (inclusive).
- user_token (string) required secret: Your User Token. See <a href="https://docs.sonarcloud.io/advanced-setup/user-accounts/">here</a>. The token is case sensitive.
- secret fields: user_token

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
pm connectors inspect source-sonar-cloud
```

### Inspect as JSON

```bash
pm connectors inspect source-sonar-cloud --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [SonarCloud Web API](https://sonarcloud.io/web_api)
- [SonarCloud authentication](https://docs.sonarsource.com/sonarcloud/advanced-setup/user-accounts/)
