# PRD Coverage

Source PRD: `POLYMETRICS_GO_CLI_MONOLITH_PRD_ARCHITECTURE.md`

Covered in this MVP:

- CLI command surface for init, help, connectors, credentials, connections, catalog, ETL, query, reverse ETL, docs, and agent planning.
- Local encrypted credential storage.
- Connector contract and registry.
- Local sample/file/warehouse/outbox connectors.
- ETL and reverse ETL workflows.
- Agent-oriented JSON output.

Not covered in this MVP:

- SQLite metadata store.
- DuckDB query engine.
- OS keychain integration.
- Parquet batches.
- SaaS connector APIs.
- Background scheduler daemon.

Reason: those require dependency choices or external systems. The first slice stays dependency-free and runnable.

