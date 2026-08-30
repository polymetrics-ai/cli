# Plan — issue 4293 source-operation multi-lane manifest

1. Add focused red tests for a source-lock-backed fixture: duplicate source ID, source row absent from the manifest, pageable row without an ETL cell, and artifact link to an unknown cell. Include all seven lanes and each allowed cell state in the passing fixture.
2. Add the smallest closed mapping-manifest model and structural validation beside existing declaration-admission schemas. Reuse the existing mapping-only source-lock reader; do not change source import, runtime, or certification code.
3. Add a `connectorgen` authoring checker that validates selected source locks, source citations, classification-fact citations, one source row per locked operation, typed state reasons, and artifact-to-cell links. Permit multiple locks for one connector and require an explicit same-identity canonical representative before source rows are deduplicated for a canonical-operation count.
4. Refactor for deterministic diagnostic sorting and focused helpers, add developer-command help plus one canonical-doc section, run `gofmt`, then verify focused Go tests and applicable existing source/map checks.

## Slice boundaries

- Included: mapping metadata, source-lock reconciliation, explicit supplemental-source lineage, static JSON schema validation, deterministic `--check` behavior, focused synthetic fixtures/tests, developer-command usage text, one canonical authoring-doc section, and planning evidence.
- Excluded: connector-specific mapping population, source lock mutation, importer changes, executor/runtime admission, ETL/reverse-ETL/sync execution, credentials, provider I/O, certification behavior, generated website data, and broad documentation generation.

## Expected red / green / refactor evidence

- Red: the new tests fail because no mapping-manifest command, model, source-lock reconciliation, or semantic diagnostics exist.
- Green: all focused cases pass, including the independent duplicate, missing-row, missing-ETL, orphan-artifact, typed-reason, supplemental-source-lineage, and incompatible-canonical-relation mutations.
- Refactor: diagnostics are stable and sorted; `gofmt` and existing authoring checks remain green without runtime changes.

## Atlas disposition

Reuse `source.projection-admission.v1` (owner symbols `parseDeclarationAdmissionSourceLock`, `runDeclarationAdmission`, and `buildOperationEvidenceForConnectors`). No shared runtime foundation is proposed, so no Atlas catalog change or captain approval is needed.

## CLI parity disposition

- `connectorgen` developer help: applicable and updated in `cmd/connectorgen/main.go`; the focused suite covers `source-operation-mapping --help`.
- `pm` runtime help, bare namespaces, `docs/cli/**`, website docs, generated PM manuals, and completions: not applicable. The checker is a source-authoring utility with no production `pm` command, no generated PM surface, no JSON mode, and no runtime behavior.
