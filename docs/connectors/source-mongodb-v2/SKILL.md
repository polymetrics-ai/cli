---
name: pm-source-mongodb-v2
description: MongoDb connector knowledge and safe action guide.
---

# pm-source-mongodb-v2

## Purpose

MongoDb catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/mongodb.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://www.mongodb.com/docs/

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: database_go
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

- family: database_cdc_source
- priority_wave: 1
- etl_operations: catalog, check, read_cdc, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- cdc_modes: snapshot, mongodb_change_streams
- cdc_state_fields: resume_token, cluster_time, snapshot_completed
- conformance: catalog, cdc_checkpoint, cdc_setup_validation, check, delete_semantics, docs_skill, ordering, read_fixture, secret_redaction, snapshot_consistency, spec, state_checkpoint

## Official Application Documentation

- MongoDB documentation: https://www.mongodb.com/docs/
- MongoDB authentication: https://www.mongodb.com/docs/manual/core/authentication/
- Release Notes: https://www.mongodb.com/docs/manual/release-notes/

## Configuration

- database_config (object) required: Configures the MongoDB cluster type.
- discover_sample_size (integer): The maximum number of documents to sample when attempting to discover the unique fields for a collection.
- discover_timeout_seconds (integer): The amount of time the connector will wait when it discovers a document. Defaults to 600 seconds. Valid range: 5 seconds to 1200 seconds.
- initial_load_timeout_hours (integer): The amount of time an initial load is allowed to continue for before catching up on CDC logs.
- initial_waiting_seconds (integer): The amount of time the connector will wait when it launches to determine if there is new data to sync or not. Defaults to 300 seconds. Valid range: 120 seconds to 1200 seconds.
- invalid_cdc_cursor_position_behavior (string): manual intervention needed
- queue_size (integer): The size of the internal queue. This may interfere with memory consumption and efficiency of the connector, please be careful.
- update_capture_mode (string): manual intervention needed
- secret fields: database_config.password

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
pm connectors inspect source-mongodb-v2
```

### Inspect as JSON

```bash
pm connectors inspect source-mongodb-v2 --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [MongoDB documentation](https://www.mongodb.com/docs/)
- [MongoDB authentication](https://www.mongodb.com/docs/manual/core/authentication/)
- [Release Notes](https://www.mongodb.com/docs/manual/release-notes/)
