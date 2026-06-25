# Native Runtime Capability Matrix PRD

## Goal

Expose check, catalog, read, write, query, ETL, and reverse ETL support for every connector in the native Go connector catalog without pretending that planned native ports are runnable.

## Problem

The catalog lists 647 connectors, but only built-in native Go connectors should be executable. Agents need a deterministic way to discover which operations are safe, which are unavailable, and why.

## Requirements

- Every `ConnectorDefinition` includes `runtime_capabilities`.
- `runtime_capabilities` exposes metadata, check, catalog, read, write, query, etl, reverse_etl, and unsupported reason.
- Enabled native connectors advertise only operations that the native implementation supports.
- Planned native ports advertise metadata-only support and a clear unsupported reason.
- CLI JSON and manual output show the capability matrix.
- Generated catalog JSON, markdown, connector manuals, and connector skills include the same capability information.
- Tests fail if any catalog connector lacks capability metadata.

## Non-Goals

- Do not enable all 647 connectors for data-plane execution in this phase.
- Do not run upstream connector images.
- Do not add Go module dependencies.
