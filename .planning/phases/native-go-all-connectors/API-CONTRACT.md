# API Contract

## Connector Interfaces

- `Connector`: existing check, catalog, read, write interface.
- `WriteValidator`: validates write payloads before reverse ETL plans.
- `DryRunWriter`: returns write previews without mutation.
- `Querier`: executes bounded SELECT-only query requests.
- `CDCReader`: emits CDC events with checkpoint state.
- `StatefulReader`: creates initial stream state.
- `SchemaMapper`: maps connector schemas to runtime stream schemas.
- `LiveConformanceProvider`: optional provider for live sandbox conformance config.

## CLI Contracts

- Direct ETL envelopes: `ETLCheck`, `ETLCatalog`, `ETLRead`.
- Catalog envelopes remain `ConnectorCatalog`, `ConnectorDefinition`, and `NativePortPlan`.
- Reverse ETL mutation contract remains plan, preview, approval token, run, and receipt.
