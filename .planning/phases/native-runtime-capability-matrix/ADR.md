# ADR: Explicit Runtime Capabilities For Every Catalog Connector

## Decision

Add a generated `runtime_capabilities` object to every catalog connector and keep planned native ports metadata-only until a native Go implementation passes connector conformance tests.

## Rationale

Agents must not infer runtime support from catalog presence. A capability matrix is safer than registering placeholder connectors because it prevents accidental credential prompts, check calls, ETL runs, or reverse ETL writes against unsupported implementations.

## Consequences

- All connectors are discoverable immediately.
- Enabled connectors remain honest about supported operations.
- Planned connectors can be prioritized and scaffolded without exposing unsafe runtimes.
