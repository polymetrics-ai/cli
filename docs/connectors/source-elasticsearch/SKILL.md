---
name: pm-source-elasticsearch
description: Elasticsearch connector knowledge and safe action guide.
---

# pm-source-elasticsearch

## Purpose

Elasticsearch catalog connector for https://docs.airbyte.com/integrations/sources/elasticsearch. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: native_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-elasticsearch:0.1.5 (metadata only; not executed)

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

- Elasticsearch REST APIs: https://www.elastic.co/guide/en/elasticsearch/reference/current/rest-apis.html
- Elasticsearch authentication: https://www.elastic.co/guide/en/elasticsearch/reference/current/setting-up-authentication.html
- Elasticsearch Release Notes: https://www.elastic.co/docs/release-notes/elasticsearch
- Elastic Cloud Status: https://status.elastic.co/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/elasticsearch

## Configuration

- authenticationMethod (object): The type of authentication to be used
- endpoint (string) required: The full url of the Elasticsearch server
- secret fields: authenticationMethod.apiKeySecret, authenticationMethod.password

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/elasticsearch

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-elasticsearch
```

### Inspect as JSON

```bash
pm connectors inspect source-elasticsearch --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Elasticsearch documentation](https://docs.airbyte.com/integrations/sources/elasticsearch)
