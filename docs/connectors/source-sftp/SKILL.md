---
name: pm-source-sftp
description: SFTP connector knowledge and safe action guide.
---

# pm-source-sftp

## Purpose

SFTP catalog connector for https://docs.airbyte.com/integrations/sources/sftp. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: file_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-sftp:0.2.4 (metadata only; not executed)

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

- family: file_object_source
- priority_wave: 3
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: bounded_streaming, catalog, check, docs_skill, format_detection, path_safety, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- No upstream application documentation URL was listed in the imported connector registry.
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/sftp

## Configuration

- credentials (object): The server authentication method
- file_pattern (string): The regular expression to specify files for sync in a chosen Folder Path
- file_types (string): Coma separated file types. Currently only 'csv' and 'json' types are supported.
- folder_path (string): The directory to search files for sync
- host (string) required: The server host address
- port (integer) required: The server port
- user (string) required: The server user
- secret fields: credentials.auth_ssh_key, credentials.auth_user_password

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/sftp

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-sftp
```

### Inspect as JSON

```bash
pm connectors inspect source-sftp --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [SFTP documentation](https://docs.airbyte.com/integrations/sources/sftp)
