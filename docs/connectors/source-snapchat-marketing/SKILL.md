---
name: pm-source-snapchat-marketing
description: Snapchat Marketing connector knowledge and safe action guide.
---

# pm-source-snapchat-marketing

## Purpose

Snapchat Marketing catalog connector for https://docs.airbyte.com/integrations/sources/snapchat-marketing. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-snapchat-marketing:1.5.40 (metadata only; not executed)

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

- Ads API announcements: https://developers.snap.com/api/marketing-api/Ads-API/announcements
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/snapchat-marketing

## Configuration

- action_report_time (string): Specifies the principle for conversion reporting.
- ad_account_ids (array): Ad Account IDs of the ad accounts to retrieve
- client_id (string) required secret: The Client ID of your Snapchat developer application.
- client_secret (string) required secret: The Client Secret of your Snapchat developer application.
- end_date (string): Date in the format 2017-01-25. Any data after this date will not be replicated.
- organization_ids (array): The IDs of the organizations to retrieve
- refresh_token (string) required secret: Refresh Token to renew the expired Access Token.
- start_date (string): Date in the format 2022-01-01. Any data before this date will not be replicated.
- swipe_up_attribution_window (string): Attribution window for swipe ups.
- view_attribution_window (string): Attribution window for views.
- secret fields: client_id, client_secret, refresh_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/snapchat-marketing

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-snapchat-marketing
```

### Inspect as JSON

```bash
pm connectors inspect source-snapchat-marketing --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Snapchat Marketing documentation](https://docs.airbyte.com/integrations/sources/snapchat-marketing)
