---
name: pm-source-hubspot
description: HubSpot connector knowledge and safe action guide.
---

# pm-source-hubspot

## Purpose

HubSpot catalog connector for https://docs.airbyte.com/integrations/sources/hubspot. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-hubspot:6.7.0 (metadata only; not executed)

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

- HubSpot API reference: https://developers.hubspot.com/docs/api/overview
- HubSpot authentication: https://developers.hubspot.com/docs/api/oauth-quickstart-guide
- HubSpot API changelog: https://developers.hubspot.com/changelog
- HubSpot rate limits: https://developers.hubspot.com/docs/api/usage-details
- HubSpot Status: https://status.hubspot.com/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/hubspot

## Configuration

- association_streams (array)
- credentials (object) required: Choose how to authenticate to HubSpot.
- custom_object_association_streams (array)
- enable_experimental_streams (boolean): If enabled then experimental streams become available for sync.
- lookback_window (integer): How far back (in minutes) to re-fetch records during incremental syncs for CRM Search streams (e.g. contacts, companies, deals, tickets). Set this if you notice missing records ...
- num_worker (integer): The number of worker threads to use for the sync.
- property_history_lookback_window (integer): How far back (in minutes) to re-fetch records during incremental syncs for property history streams (deals, contacts, companies property history). Set this if you notice missing...
- start_date (string): UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated. If not set, "2006-06-01T00:00:00Z" (Hubspot creation date) will be used a...
- treat_numbers_and_booleans_as_strings (boolean): If enabled, HubSpot dynamic `number` and `boolean` properties are exposed as `string`. Useful when HubSpot returns values that do not match the declared type and the destination...
- secret fields: credentials.access_token, credentials.client_secret, credentials.refresh_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/hubspot

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-hubspot
```

### Inspect as JSON

```bash
pm connectors inspect source-hubspot --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [HubSpot documentation](https://docs.airbyte.com/integrations/sources/hubspot)
