# TDD ledger — Issue 4331

| Slice | Red evidence | Green evidence | Refactor / regression evidence | Status |
| --- | --- | --- | --- | --- |
| Existing batch failure | `go run ./cmd/connectorgen validate /tmp/extracted-batch67/internal/connectors/defs` exited 1 with 20 batch source-projection parse failures, including 17 `unknown field "source_kind"` failures | The same command reports no rendered-reference contract failures | N/A | red recorded |
| Rendered document projection | The expanded focused suite exited 1 before implementation: every new lock was rejected by strict decode (`unknown field "coverage_confidence"`) | `go test -timeout 20m ./cmd/connectorgen -run '^TestSourceImportVersion3(RenderedReferenceProjectsCapturedCitation|MixedOpenAPIAndRenderedReferenceKeepsOpenAPIProjectionBytes|RenderedReferenceRejectsUnverifiableEvidenceAndCitations|BundleRejectsArchiveHashMismatch|UnavailableSourceProjectsBlockingGap)$|^TestSourceImportRenderedReferenceKeepsSchemaOneAndTwoLocksValid$' -count=1` passes | The established OpenAPI-only projection has fixed SHA-256 `b4f02389465992c5daf69c2a98e989058c449cc8a54fa892a016e2b9d865e4e7`; the mixed-lock test compares the bytes directly. | green recorded |
| Integrity and citations | The red suite could not decode captured evidence, citations, confidence, bundle, or unavailable fields | The green suite rejects captured rendered/bundle hash mismatches, missing/foreign citations, and an empty confidence basis | Schema 1/2 and OpenAPI-only fixtures retain existing results | green recorded |

No-fetch invariant: source import validates locally captured evidence only; cited URLs remain provenance metadata and are never fetched.
