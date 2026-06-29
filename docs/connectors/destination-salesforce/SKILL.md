---
name: pm-destination-salesforce
description: Salesforce connector knowledge and safe action guide.
---

# pm-destination-salesforce

## Purpose

Salesforce catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/salesforce.svg
- source: official
- review_status: official_verified
- review_url: https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/intro_rest.htm

## Capabilities

- catalog_metadata=true
- connector type: destination
- release stage: alpha
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: destination_go
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

- family: destination_writer
- priority_wave: 3
- etl_operations: catalog, check, write_append, write_dedup, write_overwrite
- reverse_etl_operations: none until native write conformance passes
- conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

## Official Application Documentation

- Salesforce documentation: https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/intro_rest.htm

## Configuration

- auth_type (string) required
- client_id (string) required: Enter your Salesforce developer application's <a href="https://developer.salesforce.com/forums/?id=9062I000000DLgbQAG">Client ID</a>.
- client_secret (string) required secret: Enter your Salesforce developer application's <a href="https://developer.salesforce.com/forums/?id=9062I000000DLgbQAG">Client secret</a>.
- is_sandbox (boolean) required: Toggle if you're using a <a href="https://help.salesforce.com/s/articleView?id=sf.deploy_sandboxes_parent.htm&type=5">Salesforce Sandbox</a>.
- object_storage_config (object)
- refresh_token (string) required secret: manual intervention needed
- secret fields: client_secret, object_storage_config.access_key_id, object_storage_config.secret_access_key, refresh_token

## Sync Modes

- supported sync modes: append
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Commands

### Inspect catalog entry

```bash
pm connectors inspect destination-salesforce
```

### Inspect as JSON

```bash
pm connectors inspect destination-salesforce --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Salesforce documentation](https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/intro_rest.htm)
