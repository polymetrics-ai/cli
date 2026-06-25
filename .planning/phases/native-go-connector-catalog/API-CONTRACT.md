# Native Go Connector Catalog API Contract

The public CLI API is:

```text
pm connectors list [--all] [--json]
pm connectors catalog [--type source|destination] [--stage alpha|beta|generally_available] [--json]
pm connectors inspect <name-or-catalog-slug> [--json]
```

JSON envelopes:

- `ConnectorList`: existing built-in connector list.
- `ConnectorCatalog`: all or filtered generated catalog definitions.
- `ConnectorDefinition`: one catalog-only connector definition.

Secrets are represented by field names only.
