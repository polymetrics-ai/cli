# Native Go All Connectors PLAN

## Implementation Steps

1. Add red tests requiring all 647 catalog entries to be enabled, registered, conformant, documented, and runnable through fixture operations.
2. Add optional connector interfaces for query, dry-run write, CDC, state, schema mapping, and live conformance.
3. Implement `NativeCatalogConnector` as the shared native binding for catalog slugs.
4. Derive runtime capabilities and enabled status in the catalog loader.
5. Register all native catalog connectors in `NewRegistry()`.
6. Add native conformance reports and fixture benchmark hooks.
7. Add direct `pm etl check/catalog/read` commands.
8. Regenerate connector docs and validate them.
9. Run local verification.
