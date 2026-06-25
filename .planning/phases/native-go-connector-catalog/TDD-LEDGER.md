# TDD Ledger

Phase: native-go-connector-catalog

Record failing test evidence before production code for every behavior-adding task.

## Red Evidence

### Catalog model and filters

Command:

```bash
go test ./internal/connectors -run TestConnectorCatalog
```

Expected failure:

```text
undefined: ConnectorCatalog
undefined: ConnectorCatalogCounts
undefined: ConnectorDefinitionBySlug
undefined: ImplementationEnabled
undefined: RuntimeNativeGo
```

### CLI catalog discovery and docs

Command:

```bash
go test ./internal/cli -run TestConnectorCatalog
```

Expected failure:

```text
catalog json missing "\"kind\": \"ConnectorCatalog\""
Run(connectors catalog) code = 2 stderr = error: invalid usage
```

## Green Evidence

### Catalog model and filters

Command:

```bash
go test ./internal/connectors -run TestConnectorCatalog
```

Result:

```text
ok  	polymetrics/internal/connectors
```

### CLI catalog discovery and docs

Command:

```bash
go test ./internal/cli -run TestConnectorCatalog
```

Result:

```text
ok  	polymetrics/internal/cli
```
