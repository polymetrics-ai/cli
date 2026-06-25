# Threat Model

## Risks

- Secret leakage through generated docs, JSON output, logs, or test fixtures.
- Unsafe generic external writes.
- Unsafe SQL execution through query-capable connectors.
- Untrusted connector runtime execution.

## Controls

- Generic native connectors never execute connector images or non-Go code.
- Secret fields are documented by name only; values are not rendered.
- Generic writes produce local receipts unless a connector-specific adapter implements approved external actions.
- Query accepts SELECT-only statements and rejects mutation tokens.
- Reverse ETL writes require preview and approval.
