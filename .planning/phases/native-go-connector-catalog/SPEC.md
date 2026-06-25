# Native Go Connector Catalog SPEC

## Catalog Definition

Each catalog entry must include slug, name, type, documentation URL, release stage, support level, source type, language, tags, config schema, secret fields, supported sync modes, implementation status, runtime kind, upstream image reference, and optional native `pm` connector name.

## CLI Contract

- `pm connectors list --all --json`
  - Returns `kind=ConnectorCatalog`, `count`, and `connectors`.
- `pm connectors catalog [--type source|destination] [--stage stage] [--json]`
  - Returns the same catalog envelope after filters.
- `pm connectors inspect <slug> --json`
  - For built-in connectors: existing manifest response.
  - For catalog-only connectors: `kind=ConnectorDefinition` with the catalog definition.
- `pm connectors inspect <slug>`
  - For catalog-only connectors: man-style manual using the existing guide renderer.

## Native Runtime Policy

Catalog entries are metadata-first. `implementation_status=planned_native_port` means the connector is discoverable but not runnable. No Docker/Podman connector runtime is allowed for catalog execution.
