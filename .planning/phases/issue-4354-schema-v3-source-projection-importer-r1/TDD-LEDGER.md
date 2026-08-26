# Schema-v3 source projection/importer foundation — TDD ledger

## Planned red/green/refactor slices

| Slice | Red assertion before production code | Green assertion | Refactor guard |
| --- | --- | --- | --- |
| Source-reference admission | The cited-only Outreach input fails in the current byte-backed reader/importer path. | An explicit declaration-only result preserves source URL, source operation identity, and declaration identity without claiming byte import. | No network/cache fallback, re-pin, descriptor identity rewrite, or weakening of byte-backed verification. |
| Shared reader proof | A second independent schema-v3 fixture fails at the same absent source-reference boundary. | It produces a distinct cited operation descriptor through the same common code. | Unsupported/malformed kinds still fail with contextual errors and size/path bounds. |
| Six-lane projection | A recognized unsupported operation is omitted, incorrectly blocked as provenance, or promoted to `implemented`. | It has six visible lane classifications and a specific `missing_foundation` reason. | Existing executable operations keep their canonical mapping and preflight remains the implementation authority. |

## Red

2026-08-26 — executed before production edits:

```text
go test -timeout 20m ./cmd/connectorgen -run 'TestSourceImport(OutreachReferenceProjectsCitedOperationsWithoutFetching|V3SourceReferenceUsesTheSameClosedProjection|SourceReferenceRejectsUnsupportedAndUnsafeKinds|ReferenceFeedsOperationEvidenceWithoutEnabledClassification)$' -count=1
```

Failed as intended at the closed-reader boundary:

- `TestSourceImportOutreachReferenceProjectsCitedOperationsWithoutFetching`:
  `parse source lock: json: unknown field "operation_counts"`.
- `TestSourceImportV3SourceReferenceUsesTheSameClosedProjection`:
  `parse source lock: json: unknown field "source_reference"`.
- malformed-source cases cannot reach their intended validators because the
  parser rejects the new input shape first.
- operation evidence attributes the Custom Objects URL to the primary OpenAPI
  hash/byte record and has no `source_contract_unavailable` classification.

This proves the reader has no explicit cited-only source-reference path and
would otherwise conflate an unavailable raw contract with byte-backed import.

## Green

2026-08-26 — the red command is green after adding the explicit
`source_reference` document kind and the narrow retained-Outreach legacy
adapter. It proves all of the following without a network call:

- the Outreach OpenAPI and Custom Objects operation citations retain their
  distinct URL/digest/byte identities;
- the same v3 reader path accepts a non-Outreach source-reference fixture;
- unsupported source kinds and unsafe citation URLs still fail strict parsing;
- source-import writes and checks the descriptor through its normal CLI path
  with no retained-artifact directory and reports `writes=0 cli=0`;
- source projection does not alter a matching implemented write/command; and
- operation evidence emits all six lane classifications as declared but not
  enabled, with `source_contract_unavailable` rather than a hash,
  certification, or credential failure.

```text
go test -timeout 20m ./cmd/connectorgen -run 'Test(SourceImport(OutreachReferenceProjectsCitedOperationsWithoutFetching|V3SourceReferenceUsesTheSameClosedProjection|ReferenceDigestIsProvenanceNotAnExecutionGate|SourceReferenceRejectsUnsupportedAndUnsafeKinds|ReferenceFeedsOperationEvidenceWithoutEnabledClassification)|RunSourceImportReferenceChecksWithoutRetainedArtifactOrSurfaceWrite|SourceReferenceProjectionDoesNotMaterializeAnExistingWriteOrCommand)$' -count=1
PASS

go test -timeout 20m ./cmd/connectorgen -run 'TestSourceImport(RetainedArtifactRejectsMissingAndMismatchedCopies|RejectsArtifactDriftAndSizeBeforeParsing|Version3RenderedReferenceProjectsCapturedCitation|RejectsUnsafeArtifactDestinations)$' -count=1
PASS
```

The second command retains the strict byte/digest identity and captured
cross-origin citation regressions for byte-backed import; the new path does not
reuse or weaken it.

## Refactor / review

Pending. Record error-wrapping, defensive-copy, ordering, provenance, and
generic-escape review outcomes after the final implementation.

## Verify repair red

2026-08-26 — GitHub Verify runs `32978972653` and `32978973162` both failed
at the same source-import help contract. Local reproduction:

```text
go test -timeout 20m ./cmd/connectorgen -run '^TestSourceImportCommandContractAndMigrationDocumentation$' -count=1
FAIL: source-import help is incomplete
```

The new correct help text wraps the historical literal `retained artifact`
across a line boundary and explicitly introduces the safe declaration-only
source-reference path. The assertion must move from fragile layout-dependent
text to the real contract: byte-backed retained documents remain strict and
the reference path emits `source_contract_unavailable`; no importer behavior
will change for this repair.

## Verify repair green

2026-08-26 — the assertion now checks the explicit closed contracts instead of
line layout. It requires byte-backed provenance language, the declaration-only
reference path, its precise `source_contract_unavailable` gap, and retained
content-addressed document behavior while continuing to refuse arbitrary
request/cache flags.

```text
go test -timeout 20m ./cmd/connectorgen -run '^TestSourceImportCommandContractAndMigrationDocumentation$' -count=1
PASS

go test -timeout 20m ./cmd/connectorgen -count=1
PASS (146.529s)

go vet ./cmd/connectorgen
PASS

git diff --check
PASS
```
