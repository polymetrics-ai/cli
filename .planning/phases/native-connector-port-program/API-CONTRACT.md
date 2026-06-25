# API Contract

## NativePortPlan

```json
{
  "slug": "source-postgres",
  "name": "Postgres",
  "type": "source",
  "family": "database_cdc_source",
  "runtime_kind": "database_go",
  "implementation_status": "planned_native_port",
  "priority_wave": 1,
  "etl_operations": ["check", "catalog", "read_snapshot", "read_incremental"],
  "reverse_etl_operations": [],
  "cdc": {
    "supported": true,
    "modes": ["snapshot", "postgres_logical_replication"],
    "requirements": ["wal_level=logical", "replication slot", "publication"]
  },
  "conformance": ["spec", "check", "catalog", "read", "state", "secret_redaction"]
}
```

## CLI

- `pm connectors port-plan --all [--json]`
- `pm connectors port-plan <slug> [--json]`
