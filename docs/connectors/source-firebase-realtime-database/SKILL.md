---
name: pm-source-firebase-realtime-database
description: Firebase Realtime Database connector knowledge and safe action guide.
---

# pm-source-firebase-realtime-database

## Purpose

Firebase Realtime Database catalog connector for https://docs.airbyte.com/integrations/sources/firebase-realtime-database. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: database_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-firebase-realtime-database:0.1.48 (metadata only; not executed)

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

- family: database_source
- priority_wave: 3
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, cursor_incremental, docs_skill, query_safety, read_fixture, secret_redaction, spec, state_checkpoint, type_mapping

## Official Application Documentation

- Firebase Realtime Database REST API: https://firebase.google.com/docs/reference/rest/database
- Firebase authentication: https://firebase.google.com/docs/database/rest/auth
- Firebase Status Dashboard: https://status.firebase.google.com/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/firebase-realtime-database

## Configuration

- buffer_size (number): Number of records to fetch at once
- database_name (string) required: Database name (This will be part of the url pointing to the database, https://<database_name>.firebaseio.com/)
- google_application_credentials (string) required secret: Cert credentials in JSON format of Service Account with Firebase Realtime Database Viewer role. (see, https://firebase.google.com/docs/projects/iam/roles-predefined-product#real...
- path (string): Path to a node in the Firebase realtime database
- secret fields: google_application_credentials

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/firebase-realtime-database

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-firebase-realtime-database
```

### Inspect as JSON

```bash
pm connectors inspect source-firebase-realtime-database --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Firebase Realtime Database documentation](https://docs.airbyte.com/integrations/sources/firebase-realtime-database)
