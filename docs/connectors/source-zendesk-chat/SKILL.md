---
name: pm-source-zendesk-chat
description: Zendesk Chat connector knowledge and safe action guide.
---

# pm-source-zendesk-chat

## Purpose

Zendesk Chat catalog connector for https://docs.airbyte.com/integrations/sources/zendesk-chat. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-zendesk-chat:1.3.15 (metadata only; not executed)

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

- Changelog: https://developer.zendesk.com/api-reference/changelog/changelog/
- Developer Updates: https://support.zendesk.com/hc/en-us/sections/4405298889242-Developer-updates
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/zendesk-chat

## Configuration

- credentials (object)
- start_date (string) required: The date from which you'd like to replicate data for Zendesk Chat API, in the format YYYY-MM-DDT00:00:00Z.
- subdomain (string) required: The unique subdomain of your Zendesk account (without https://). <a href=\"https://support.zendesk.com/hc/en-us/articles/4409381383578-Where-can-I-find-my-Zendesk-subdomain\">Se...
- secret fields: credentials.access_token, credentials.client_id, credentials.client_secret, credentials.refresh_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/zendesk-chat

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-zendesk-chat
```

### Inspect as JSON

```bash
pm connectors inspect source-zendesk-chat --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Zendesk Chat documentation](https://docs.airbyte.com/integrations/sources/zendesk-chat)
