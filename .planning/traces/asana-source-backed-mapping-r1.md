# Asana source-backed mapping and multipart coverage R1

## Task delivery header

- Base: `cb365f8379b8b9e421051321a9bb6c99d11301a1`.
- Code commit: `1f79bb3058d3d927785dba4e1a002f781e6c5f59`.
- Delivery: local commits only; no push, PR creation, merge, runtime change, or source-lock rewrite in this commit.
- Scope: retained Asana source-import identity, overlapping direct/ETL and direct-write/reverse-ETL evidence, and closed multipart coverage.

## Root cause and boundary

The legacy v2 Asana lock retained the provider REST inventory but the importer
looked only at a v3 `source_documents` inventory. It therefore did not replace
the artifact-derived fallback identity with the locked `asana.rest.*` identity.
That caused the custom-field, fan-out section, and partial-mutation projection
rows to appear absent even though their pinned source operations were present.

`createAttachmentForObject` was a separate mapping defect, not an execution
foundation gap. The source-projection completeness predicate only understood
JSON contracts and rejected the retained `multipart/form-data` request, while
the Asana definition already contained two tested, closed multipart actions.
The repair admits only a declared multipart action whose provider media,
part-to-source-field mapping, closed record schema, required fields, and CLI
flags all agree. It does not add a generic multipart escape hatch.

## Red / green evidence

| Case | Red result | Green proof |
| --- | --- | --- |
| Retained legacy REST identity | `asana.rest.getCustomFieldsForWorkspace`, `asana.rest.createCustomField`, and `asana.rest.getSectionsForProject` were not retained through import | `TestSourceImportRetainedAsanaPreservesLockedRESTOperationIDs` asserts all three exact locked IDs and rejects a mismatched provider operation ID before propagation. |
| Direct/read and ETL overlap | An implemented direct read with a stream lacked ETL classification | Table tests prove a direct read keeps `direct_read` and gains ETL only with an implemented stream; a binary command with a stream does not gain ETL. |
| Direct/write and reverse ETL overlap | `reverse-ETL classification = {Declared:false Enabled:false}, want {Declared:true Enabled:true}` | Table tests prove an implemented direct write keeps `direct_write` and gains reverse ETL only with an implemented write; a reverse-ETL command does not fabricate direct-write classification. |
| Asana attachment mapping | `upload_attachment_file` did not cover the locked provider operation | `TestRetainedAsanaMultipartActionsCoverLockedAttachmentOperation` accepts both closed Asana variants and rejects wrong media, missing provider-required part, a non-required provider-required part, unmapped record property, unknown provider part, and a missing required CLI flag. |

## Commands

- `gofmt -w cmd/connectorgen/operationevidence.go cmd/connectorgen/operationevidence_test.go cmd/connectorgen/sourceprojection.go cmd/connectorgen/sourceprojection_test.go`
- `go test ./cmd/connectorgen -run '^(TestSourceImportRetainedAsanaPreservesLockedRESTOperationIDs|TestOperationEvidenceClassifyKeepsImplementedStreamBoundDirectReadsInBothLanes|TestOperationEvidenceClassifyKeepsImplementedWriteBoundDirectWritesInBothLanes|TestRetainedAsanaMultipartActionsCoverLockedAttachmentOperation)$' -count=1 -v` — pass.
- `go test ./cmd/connectorgen -count=1` — pass.
- `git diff --check` — pass before commit.

## Non-goals

- No source-lock, descriptor, connector-definition, runtime HTTP, or global generated-artifact change is in code commit `1f79bb3058d3d927785dba4e1a002f781e6c5f59`.
- Strict source import still validates the retained artifact, hash, byte count, provider ID, route, and source location. The legacy inventory is an additional identity binding, not a permissive fallback.
- Certification remains independent of this mapping admission and is not consulted by the multipart predicate.
