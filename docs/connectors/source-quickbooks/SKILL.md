---
name: pm-source-quickbooks
description: QuickBooks connector knowledge and safe action guide.
---

# pm-source-quickbooks

## Purpose

QuickBooks catalog connector for https://docs.airbyte.com/integrations/sources/quickbooks. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-quickbooks:4.0.4 (metadata only; not executed)

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

- QuickBooks Online API: https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/account
- QuickBooks authentication: https://developer.intuit.com/app/developer/qbo/docs/develop/authentication-and-authorization
- QuickBooks Online API Changelog: https://developer.intuit.com/app/developer/qbo/docs/changelog
- QuickBooks rate limits: https://developer.intuit.com/app/developer/qbo/docs/develop/troubleshooting/error-codes#rate-limits
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/quickbooks

## Configuration

- access_token (string) required secret: Access token for making authenticated requests.
- auth_type (string)
- client_id (string) required: Identifies which app is making the request. Obtain this value from the Keys tab on the app profile via My Apps on the developer site. There are two versions of this key: develop...
- client_secret (string) required secret: Obtain this value from the Keys tab on the app profile via My Apps on the developer site. There are two versions of this key: development and production.
- realm_id (string) required secret: Labeled Company ID. The Make API Calls panel is populated with the realm id and the current access token.
- refresh_token (string) required secret: A token used when refreshing the access token.
- sandbox (boolean) required: Determines whether to use the sandbox or production environment.
- start_date (string) required: The default value to use if no bookmark exists for an endpoint (rfc3339 date string). E.g, 2021-03-20T00:00:00Z. Any data before this date will not be replicated.
- token_expiry_date (string) required: The date-time when the access token should be refreshed.
- secret fields: access_token, client_secret, realm_id, refresh_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/quickbooks

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-quickbooks
```

### Inspect as JSON

```bash
pm connectors inspect source-quickbooks --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [QuickBooks documentation](https://docs.airbyte.com/integrations/sources/quickbooks)
