---
name: pm-destination-sftp-json
description: SFTP-JSON connector knowledge and safe action guide.
---

# pm-destination-sftp-json

## Purpose

SFTP-JSON catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/pm-warehouse.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- catalog_metadata=true
- connector type: destination
- release stage: alpha
- support level: community

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

- manual intervention needed

## Configuration

- destination_path (string) required: Path to the directory where json files will be written.
- host (string) required: Hostname of the SFTP server.
- password (string) required secret: Password associated with the username.
- port (integer): Port of the SFTP server.
- username (string) required: Username to use to access the SFTP server.
- secret fields: password

## Sync Modes

- supported sync modes: append, overwrite
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Commands

### Inspect catalog entry

```bash
pm connectors inspect destination-sftp-json
```

### Inspect as JSON

```bash
pm connectors inspect destination-sftp-json --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.
