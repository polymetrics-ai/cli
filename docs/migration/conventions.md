# Connector execution JSON conventions

The authoritative authoring procedure is
[`SOURCE-LOCK-VNEXT.md`](../connector-canon/SOURCE-LOCK-VNEXT.md). Authors edit
schema-4 `source.lock.json`; `connectorgen lock-render` materializes these
execution files.

- `metadata.json`: identity, documentation URL, capability summary, release
  stage, and risk summary.
- `spec.json`: closed connection configuration schema; secrets use `x-secret`.
- `streams.json`: shared HTTP base and ETL stream entries.
- `schemas/*.json`: record, request, and response schemas from the lock's shared
  registry.
- `writes.json`: closed typed actions used by direct write and reverse ETL.
- `operations.json`: direct read/write, GraphQL, binary, multipart, and status
  operations consumed by registered encoders.
- `cli_surface.json`: command paths, intents, invocation flags, execution
  bindings, safety, and availability.
- `database.json`, `changefeed.json`, `polling_watermark.json`,
  `rate_limits.json`, `sync_transport.json`: optional execution contracts owned
  by their existing typed runtime paths.

Generated JSON is reviewed but not independently authored. Provider evidence
stays in the lock and does not appear in these files. A command binding must be
unambiguous, routes and schemas bounded, and every named encoder/executor real.
All saved movement crosses the connection-owned DuckDB warehouse.

```bash
go run ./cmd/connectorgen lock-render <connector>
go run ./cmd/connectorgen lock-render <connector> --check
go run ./cmd/connectorgen validate internal/connectors/defs
```
