---
name: pm-source-netsuite
description: Netsuite connector knowledge and safe action guide.
---

# pm-source-netsuite

## Purpose

Netsuite catalog connector for https://docs.airbyte.com/integrations/sources/netsuite. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: native_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-netsuite:0.1.27 (metadata only; not executed)

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

- NetSuite REST API: https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/chapter_1540391670.html
- NetSuite authentication: https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/section_4389727047.html
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/netsuite

## Configuration

- consumer_key (string) required secret: Consumer key associated with your integration
- consumer_secret (string) required secret: Consumer secret associated with your integration
- object_types (array): The API names of the Netsuite objects you want to sync. Setting this speeds up the connection setup process by limiting the number of schemas that need to be retrieved from Nets...
- realm (string) required secret: Netsuite realm e.g. 2344535, as for `production` or 2344535_SB1, as for the `sandbox`
- start_datetime (string) required: Starting point for your data replication, in format of "YYYY-MM-DDTHH:mm:ssZ"
- token_key (string) required secret: Access token key
- token_secret (string) required secret: Access token secret
- window_in_days (integer): The amount of days used to query the data with date chunks. Set smaller value, if you have lots of data.
- secret fields: consumer_key, consumer_secret, realm, token_key, token_secret

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/netsuite

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-netsuite
```

### Inspect as JSON

```bash
pm connectors inspect source-netsuite --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Netsuite documentation](https://docs.airbyte.com/integrations/sources/netsuite)
