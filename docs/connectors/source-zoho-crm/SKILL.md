---
name: pm-source-zoho-crm
description: ZohoCRM connector knowledge and safe action guide.
---

# pm-source-zoho-crm

## Purpose

ZohoCRM catalog connector for https://docs.airbyte.com/integrations/sources/zoho-crm. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: native_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-zoho-crm:0.1.3 (metadata only; not executed)

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
- priority_wave: 3
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Zoho CRM API: https://www.zoho.com/crm/developer/docs/api/v6/
- Zoho OAuth 2.0: https://www.zoho.com/crm/developer/docs/api/v6/oauth-overview.html
- Zoho CRM API changelog: https://www.zoho.com/crm/developer/docs/api/v6/whats-new.html
- Zoho API limits: https://www.zoho.com/crm/developer/docs/api/v6/api-limits.html
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/zoho-crm

## Configuration

- client_id (string) required secret: OAuth2.0 Client ID
- client_secret (string) required secret: OAuth2.0 Client Secret
- dc_region (string) required: Please choose the region of your Data Center location. More info by this <a href="https://www.zoho.com/crm/developer/docs/api/v2/multi-dc.html">Link</a>
- edition (string) required: Choose your Edition of Zoho CRM to determine API Concurrency Limits
- environment (string) required: Please choose the environment
- refresh_token (string) required secret: OAuth2.0 Refresh Token
- start_datetime ([string null]): ISO 8601, for instance: `YYYY-MM-DD`, `YYYY-MM-DD HH:MM:SS+HH:MM`
- secret fields: client_id, client_secret, refresh_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/zoho-crm

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-zoho-crm
```

### Inspect as JSON

```bash
pm connectors inspect source-zoho-crm --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [ZohoCRM documentation](https://docs.airbyte.com/integrations/sources/zoho-crm)
