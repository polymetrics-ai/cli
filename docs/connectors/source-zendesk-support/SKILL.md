---
name: pm-source-zendesk-support
description: Zendesk Support connector knowledge and safe action guide.
---

# pm-source-zendesk-support

## Purpose

Zendesk Support catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/zendesk-support.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.zendesk.com/api-reference/ticketing/introduction/

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

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
- priority_wave: 1
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Zendesk Support API: https://developer.zendesk.com/api-reference/ticketing/introduction/
- Zendesk authentication: https://developer.zendesk.com/api-reference/ticketing/introduction/#security-and-authentication
- API Changelog: https://developer.zendesk.com/api-reference/changelog/changelog/
- Zendesk API changelog: https://developer.zendesk.com/api-reference/ticketing/introduction/#changes
- Zendesk rate limits: https://developer.zendesk.com/api-reference/ticketing/account-configuration/usage_limits/
- Zendesk Status: https://status.zendesk.com/

## Configuration

- credentials (object): manual intervention needed
- ignore_pagination (boolean): [Deprecated] Makes each stream read a single page of data.
- num_workers (integer): The number of worker threads to use for the sync. Higher values can improve sync throughput on large workspaces; lower values reduce load on the source.
- page_size (integer): The number of records per page for the ticket_comments stream API requests. Lower values may help prevent timeouts on large datasets. The maximum value is 1000.
- start_date (string): The UTC date and time from which you'd like to replicate data, in the format YYYY-MM-DDT00:00:00Z. All data generated after this date will be replicated.
- subdomain (string) required: This is your unique Zendesk subdomain that can be found in your account URL. For example, in https://MY_SUBDOMAIN.zendesk.com/, MY_SUBDOMAIN is the value of your subdomain.
- secret fields: credentials.access_token, credentials.api_token, credentials.client_id, credentials.client_secret, credentials.refresh_token

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
pm connectors inspect source-zendesk-support
```

### Inspect as JSON

```bash
pm connectors inspect source-zendesk-support --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Zendesk Support API](https://developer.zendesk.com/api-reference/ticketing/introduction/)
- [Zendesk authentication](https://developer.zendesk.com/api-reference/ticketing/introduction/#security-and-authentication)
- [API Changelog](https://developer.zendesk.com/api-reference/changelog/changelog/)
- [Zendesk API changelog](https://developer.zendesk.com/api-reference/ticketing/introduction/#changes)
- [Zendesk rate limits](https://developer.zendesk.com/api-reference/ticketing/account-configuration/usage_limits/)
- [Zendesk Status](https://status.zendesk.com/)
