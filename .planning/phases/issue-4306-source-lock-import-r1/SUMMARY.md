---
coverage:
  - id: D1
    description: Closed lock-only source retrieval verifies bytes and SHA-256 before parsing or output.
    verification:
      - kind: unit
        ref: cmd/connectorgen/sourceimport_test.go: TestSourceImportCommandUsesOnlyConnectorOwnedLockAndCheckMode
        status: pass
    human_judgment: false
  - id: D2
    description: Canonical descriptors preserve exact fixed contracts and complete response status shapes for multiple connector identities.
    verification:
      - kind: unit
        ref: cmd/connectorgen/sourceimport_test.go: TestSourceImportProducesClosedCanonicalDescriptors
        status: pass
    human_judgment: false
  - id: D3
    description: Unsupported, ambiguous, unbounded, drifting, or oversized provider artifacts fail before generation.
    verification:
      - kind: unit
        ref: cmd/connectorgen/sourceimport_test.go: TestSourceImportRejectsUnsafeOrUnboundedSourceForms
        status: pass
      - kind: unit
        ref: cmd/connectorgen/sourceimport_test.go: TestSourceImportRejectsArtifactDriftAndSizeBeforeParsing
        status: pass
    human_judgment: false
  - id: D4
    description: Batches have a closed help and migration adoption contract with no generic request inputs.
    verification:
      - kind: unit
        ref: cmd/connectorgen/sourceimport_test.go: TestSourceImportCommandContractAndMigrationDocumentation
        status: pass
    human_judgment: false
---

# Summary — source-lock operation import

## Delivered

- Added `connectorgen source-import`, a closed source-artifact importer that resolves only a connector-owned source lock, fetches the lock URL without credentials, refuses redirects, and validates exact size/SHA-256 before parsing.
- Added deterministic canonical operation descriptors for OpenAPI 3 YAML and Swagger 2 JSON with fixed lower-case endpoint identity, separated request schemas, source provenance, auth/pagination/byte metadata, complete resolved response declarations, and JSON/binary/status/text classification.
- Added fail-closed local-reference, bounded-schema, identity, callback, encoding, count/depth, and artifact-drift protections; no connector definition or provider artifact was added.
- Added two synthetic connector locks/artifacts, rejection fixtures, command/check-mode coverage, and migration batch-adoption guidance.

## GSD lifecycle

Executed inline/manual fallback: discussion, TDD planning, implementation, verification, and code review were completed in this phase directory because the canonical single-worker contract prohibits spawning roles in this runtime.

## Validation

Focused source-import tests, full `cmd/connectorgen` tests, generator validation/surface sync, `go vet`, `pm` build, completion-tracked connector boundary, and the repaired full `make verify` run passed. The first full run found one unchecked HTTP body-close error; the fix was made and the full run passed.
