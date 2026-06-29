---
name: pm-source-salesforce
description: Salesforce connector knowledge and safe action guide.
---

# pm-source-salesforce

## Purpose

Salesforce catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/salesforce.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/rest_rns.htm

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: native_go
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

- family: custom_go_port
- priority_wave: 1
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- REST API Release Notes: https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/rest_rns.htm
- Winter 2026 release notes - API: https://help.salesforce.com/s/articleView?id=release-notes.salesforce_release_notes.htm&release=258&type=5

## Configuration

- auth_type (string)
- client_id (string) required: Enter your Salesforce developer application's <a href="https://developer.salesforce.com/forums/?id=9062I000000DLgbQAG">Client ID</a>
- client_secret (string) required secret: Enter your Salesforce developer application's <a href="https://developer.salesforce.com/forums/?id=9062I000000DLgbQAG">Client secret</a>
- force_use_bulk_api (boolean): Toggle to use Bulk API (this might cause empty fields for some streams)
- is_sandbox (boolean): Toggle if you're using a <a href="https://help.salesforce.com/s/articleView?id=sf.deploy_sandboxes_parent.htm&type=5">Salesforce Sandbox</a>
- lookback_window (string): manual intervention needed
- refresh_token (string) required secret: manual intervention needed
- start_date (string): manual intervention needed
- stream_slice_step (string): The size of the time window (ISO8601 duration) to slice requests.
- streams_criteria (array): Add filters to select only required stream based on `SObject` name. Use this field to filter which tables are displayed by this connector. This is useful if your Salesforce acco...
- secret fields: client_secret, refresh_token

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
pm connectors inspect source-salesforce
```

### Inspect as JSON

```bash
pm connectors inspect source-salesforce --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [REST API Release Notes](https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/rest_rns.htm)
- [Winter 2026 release notes - API](https://help.salesforce.com/s/articleView?id=release-notes.salesforce_release_notes.htm&release=258&type=5)
