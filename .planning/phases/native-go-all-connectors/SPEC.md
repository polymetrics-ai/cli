# Native Go All Connectors SPEC

## Native Binding Model

The generated catalog remains the upstream metadata source. At load time, `ConnectorCatalog()` derives native enablement by setting every definition to `implementation_status=enabled` and computing runtime capabilities from connector type and runtime family.

`NewRegistry()` registers the original hand-written connectors and then registers one `NativeCatalogConnector` for every catalog slug. Short names such as `github`, `warehouse`, `sample`, `file`, and `outbox` remain unchanged.

## Runtime Behavior

- Source connectors return a deterministic fixture stream and one fixture record unless a connector-specific implementation overrides behavior.
- Destination connectors validate approved write actions and write local JSONL receipts under `.polymetrics/native/<slug>/<table>.jsonl` when a project directory is available.
- Query-capable native connectors reject non-SELECT statements and return bounded fixture rows.
- CDC-capable native connectors emit a fixture CDC event with family-specific checkpoint state.

## CLI

`pm etl` adds direct native operations:

- `pm etl check --connector <slug>`
- `pm etl catalog --connector <slug>`
- `pm etl read --connector <slug> [--stream stream] [--limit n]`

Reverse ETL keeps the existing plan, preview, approval, and run boundary.
