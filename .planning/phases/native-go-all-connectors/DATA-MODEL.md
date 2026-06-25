# Data Model

## Runtime Metadata

`ConnectorDefinition` remains the catalog model. At runtime, the loader derives:

- `implementation_status=enabled`
- `runtime_capabilities`
- native support notes

## Receipts

Generic native destination writes emit JSONL receipts:

```text
.polymetrics/native/<connector-slug>/<table>.jsonl
```

Each receipt includes connector slug, action, and the redacted record payload passed through reverse ETL.
