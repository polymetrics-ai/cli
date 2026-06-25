# Phase Summary

Phase: native-runtime-capability-matrix

## Completed

- Added `RuntimeCapabilities` to generated connector catalog definitions.
- Populated check, catalog, read, write, query, ETL, and reverse ETL capability flags for every one of the 647 catalog connectors.
- Marked enabled `source-github` as native Go capable for check, catalog, read, write, ETL, and reverse ETL.
- Marked all planned native ports as metadata-only with a clear unsupported reason.
- Updated catalog manuals, skills, markdown, JSON docs, and CLI inspection output.
- Added tests that prevent catalog-only connectors from being mistaken for runnable connectors.
- Regenerated the embedded catalog and connector documentation.
- Installed the updated `pm` binary at `/Users/karthiksivadas/.local/bin/pm`.

## Boundary

This phase does not implement 646 native connector data planes. It implements the required safe all-connector contract so agents can discover every connector and know exactly which operations are runnable. Native ports must still be enabled connector-by-connector behind conformance tests.
