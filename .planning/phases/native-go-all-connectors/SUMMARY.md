# Phase Summary

Phase: native-go-all-connectors

## Completed

- Corrected connector implementation status handling so the embedded catalog is no longer rewritten to `enabled` at runtime.
- Preserved the real catalog split: 647 total connectors, 591 sources, 56 destinations, 1 enabled live native connector, and 646 planned native ports.
- Changed the runtime registry to expose only enabled executable connectors plus built-in local connectors.
- Added a `source-github` catalog alias that delegates to the existing live GitHub connector instead of using a generic fixture connector.
- Kept fixture-backed scaffold conformance for all 647 catalog entries as metadata/docs/secret-redaction coverage, not as live runtime enablement.
- Updated CLI tests so planned connectors cannot create credentials, run ETL, run reverse ETL, or appear as executable registry entries.
- Regenerated connector docs and catalog docs with `planned_native_port` statuses and persistent native implementation program links.

## Boundary

This does not implement live native ports for the remaining 646 connectors. It implements the enforcement layer required before those ports can be safely added: no connector is executable until its actual native Go adapter passes conformance and is marked enabled.
