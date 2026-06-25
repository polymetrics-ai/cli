---
name: pm-destination-hubspot
description: HubSpot connector knowledge and safe action guide.
---

# pm-destination-hubspot

## Purpose

HubSpot catalog connector for https://docs.airbyte.com/integrations/destinations/hubspot. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: destination
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: destination_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/destination-hubspot:0.0.12 (metadata only; not executed)

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

- family: destination_writer
- priority_wave: 3
- etl_operations: catalog, check, write_append, write_dedup, write_overwrite
- reverse_etl_operations: none until native write conformance passes
- conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

## Official Application Documentation

- HubSpot API documentation: https://developers.hubspot.com/docs/api/overview
- OAuth: https://developers.hubspot.com/docs/api/oauth-quickstart-guide
- HubSpot Developer Changelog: https://developers.hubspot.com/changelog
- Rate limits: https://developers.hubspot.com/docs/api/usage-details
- HubSpot Status: https://status.hubspot.com/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/hubspot

## Configuration

- credentials (object) required: Choose how to authenticate to HubSpot.
- object_storage_config (object)
- secret fields: credentials.client_id, credentials.client_secret, credentials.refresh_token, object_storage_config.access_key_id, object_storage_config.secret_access_key

## Sync Modes

- supported sync modes: append
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/destinations/hubspot

## Commands

### Inspect catalog entry

```bash
pm connectors inspect destination-hubspot
```

### Inspect as JSON

```bash
pm connectors inspect destination-hubspot --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [HubSpot documentation](https://docs.airbyte.com/integrations/destinations/hubspot)
