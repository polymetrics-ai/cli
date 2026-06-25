---
name: pm-source-intercom
description: Intercom connector knowledge and safe action guide.
---

# pm-source-intercom

## Purpose

Intercom catalog connector for https://docs.airbyte.com/integrations/sources/intercom. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-intercom:0.13.24 (metadata only; not executed)

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

- Unversioned Changes: https://developers.intercom.com/docs/build-an-integration/learn-more/rest-apis/unversioned-changes#unversioned-changes
- API Changelog: https://developers.intercom.com/docs/references/changelog
- Changelog: https://developers.intercom.com/docs/build-an-integration/learn-more/rest-apis/api-changelog
- Intercom API OpenAPI specification: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/openapi.json
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/intercom

## Configuration

- access_token (string) required secret: Access token for making authenticated requests. See the <a href="https://developers.intercom.com/building-apps/docs/authentication-types#how-to-get-your-access-token">Intercom d...
- activity_logs_time_step (integer): Set lower value in case of failing long running sync of Activity Logs stream.
- api_rate_limit (integer): The effective API request budget per minute. The default of 9500 is 95% of the standard Intercom rate limit (10,000/min), leaving headroom for occasional bursts. If your workspa...
- client_id (string) secret: Client Id for your Intercom application.
- client_secret (string) secret: Client Secret for your Intercom application.
- lookback_window (integer): The number of days to shift the state value backward for record sync
- num_workers (integer): The number of worker threads to use for concurrent stream processing. Increase this to speed up syncs for workspaces with large volumes of conversations. The default should work...
- start_date (string) required: UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated.
- secret fields: access_token, client_id, client_secret

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/intercom

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-intercom
```

### Inspect as JSON

```bash
pm connectors inspect source-intercom --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Intercom documentation](https://docs.airbyte.com/integrations/sources/intercom)
