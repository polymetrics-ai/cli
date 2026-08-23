---
phase: issue-4318-multidoc-source-lock-foundation-r1
status: complete
coverage:
  - id: D1
    description: A v3 source lock imports several REST artifacts with document-owned provenance and distinct locked identities even when provider operation IDs repeat.
    verification:
      - kind: integration
        ref: cmd/connectorgen/sourceimport_test.go TestSourceImportVersion3ImportsDocumentOwnedProvenanceAndDuplicateProviderIDs
        status: pass
      - kind: unit
        ref: cmd/connectorgen/sourceprojection_test.go TestSourceProjectionAcceptsVersion3DocumentProvenance
        status: pass
    human_judgment: false
  - id: D2
    description: A v3 source artifact is fetched and verified per document, rejects missing/drifted input, and deduplicates an identical digest safely.
    verification:
      - kind: integration
        ref: cmd/connectorgen/sourceimport_test.go TestSourceImportVersion3RejectsMissingOrDriftedDocument
        status: pass
      - kind: unit
        ref: cmd/connectorgen/sourceimport_test.go TestSourceImportVersion3SynchronizesDuplicateArtifactDigests (race detector)
        status: pass
    human_judgment: false
  - id: D3
    description: Existing v2 source locks and all frozen GitHub source artifacts remain byte-identical.
    verification:
      - kind: integration
        ref: go run ./cmd/connectorgen source-import github --check
        status: pass
      - kind: unit
        ref: cmd/connectorgen/sourceimport_test.go TestSourceImportPreservesFrozenGitHubArtifacts
        status: pass
    human_judgment: false
  - id: D4
    description: YAML scalar keys normalize to JSON member names without weakening duplicate detection or admitting compound keys.
    verification:
      - kind: integration
        ref: cmd/connectorgen/sourceimport_test.go TestSourceImportNormalizesScalarYAMLMappingKeys
        status: pass
    human_judgment: false
---

# SUMMARY — issue #4318 additive v3 multi-document source locks

## Delivered

- Added exact v3 source-lock decoding with a sorted aggregate OpenAPI inventory and several
  document-owned REST artifacts, while retaining v1/v2 decoding and descriptor schema 2.
- Fetches only queryless stable artifacts, verifies every document's locked digest and bytes, and
  synchronizes duplicate digest fetches. Published source URLs are retained as bounded citation
  provenance and are never fetched.
- Preserved document-qualified source identity for repeated provider `operationId` values; the
  descriptor, source projection validator, and surface sync now understand schema 3.
- Added the directly adjacent strict YAML bridge correction: standard JSON scalar mapping keys
  normalize before duplicate detection, preserving `200:`/`'200':` collision safety.

## Verification outcome

All behavioral tests, the focused race test, frozen GitHub hash test, and full `make verify` pass.
The mandatory three GitHub bytes/SHA-256 baselines are unchanged. No Zoom capture or Docker Hub
connector artifact was added.
