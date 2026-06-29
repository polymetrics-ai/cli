---
name: pm-source-smoke-test
description: Smoke Test connector knowledge and safe action guide.
---

# pm-source-smoke-test

## Purpose

Smoke Test catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

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
- priority_wave: 3
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- manual intervention needed

## Configuration

- all_fast_streams (boolean): Include all fast (non-high-volume) predefined streams.
- all_slow_streams (boolean): Include all slow (high-volume) streams such as large_batch_stream. These are excluded by default to avoid incurring the cost of large record sets.
- custom_scenarios (array): Additional test scenarios to inject at runtime. Each scenario defines a stream name, JSON schema, and records.
- large_batch_record_count (integer): Number of records to generate for the large_batch_stream scenario. Set to 0 to emit no records for this stream.
- namespace ([string null]): Namespace (schema/database) to set on all streams. When provided, the destination will write data into this namespace.
- scenario_filter (array): Specific scenario names to include. These are unioned with the boolean-driven sets (deduped). If omitted or empty, only the boolean flags control selection.

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
pm connectors inspect source-smoke-test
```

### Inspect as JSON

```bash
pm connectors inspect source-smoke-test --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.
