---
name: pm-source-jira
description: Jira connector knowledge and safe action guide.
---

# pm-source-jira

## Purpose

Jira catalog connector for https://docs.airbyte.com/integrations/sources/jira. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-jira:6.0.0 (metadata only; not executed)

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

- Changelog: https://developer.atlassian.com/changelog/#
- Jira Platform API Changelog: https://developer.atlassian.com/cloud/jira/platform/changelog/
- Jira Software API Changelog: https://developer.atlassian.com/cloud/jira/software/changelog/
- Jira Cloud Platform API OpenAPI specification: https://developer.atlassian.com/cloud/jira/platform/swagger-v3.v3.json
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/jira

## Configuration

- credentials (object) required: Choose how to authenticate to Jira.
- domain (string) required: Your Jira host (full domain — do not include 'https://' or paths). Examples: airbyteio.atlassian.net, airbyteio.jira.com, jira.your-domain.com.
- expand_issue_changelog (boolean): (DEPRECATED) Expand the changelog when replicating issues.
- expand_issue_transition (boolean): (DEPRECATED) Expand the transitions when replicating issues.
- issues_stream_expand_with (array): Select fields to Expand the `Issues` stream when replicating with:
- lookback_window_minutes (integer): When set to N, the connector will always refresh resources created within the past N minutes. By default, updated objects that are not newly created are not incrementally synced.
- num_workers (integer): The number of worker threads to use for the sync.
- projects (array): List of Jira project keys to replicate data for, or leave it empty if you want to replicate data for all projects.
- render_fields (boolean): (DEPRECATED) Render issue fields in HTML format in addition to Jira JSON-like format.
- start_date (string): The date from which you want to replicate data from Jira, use the format YYYY-MM-DDT00:00:00Z. Note that this field only applies to certain streams, and only data generated on o...
- secret fields: credentials.api_token, credentials.client_id, credentials.client_secret, credentials.refresh_token, credentials.service_account_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/jira

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-jira
```

### Inspect as JSON

```bash
pm connectors inspect source-jira --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Jira documentation](https://docs.airbyte.com/integrations/sources/jira)
