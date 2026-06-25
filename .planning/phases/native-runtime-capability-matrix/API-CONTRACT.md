# API Contract

## ConnectorDefinition

Adds:

```json
{
  "runtime_capabilities": {
    "metadata": true,
    "check": false,
    "catalog": false,
    "read": false,
    "write": false,
    "query": false,
    "etl": false,
    "reverse_etl": false,
    "unsupported_reason": "Native Go port not enabled yet."
  }
}
```

## CLI

- `pm connectors list --all --json` includes `runtime_capabilities` for all catalog connectors.
- `pm connectors catalog --json` includes `runtime_capabilities` for all returned connectors.
- `pm connectors inspect <slug> --json` includes `runtime_capabilities`.
- `pm connectors inspect <slug>` includes a `RUNTIME CAPABILITIES` manual section.
